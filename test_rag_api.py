import json
import tempfile
import threading
import unittest
from http.server import ThreadingHTTPServer
from pathlib import Path
from urllib.error import HTTPError
from urllib.request import Request, urlopen
from unittest.mock import patch

from open_coach import GeminiRequestError, OpenCoachStepError
from rag_api import (
    AskRateLimiter,
    MAX_QUESTION_CHARS,
    RAGQueryService,
    make_handler,
    parse_allowed_origins,
)
from rag_tutor import LocalRAG, NO_EVIDENCE


class FakeEmbedder:
    def __init__(self):
        self.calls = 0

    def __call__(self, text, task_type):
        self.calls += 1
        lowered = text.lower()
        return [float("pedido" in lowered), float("férias" in lowered), 0.1]


class RAGQueryServiceTests(unittest.TestCase):
    def test_rate_limit_resets_after_window(self):
        limiter = AskRateLimiter(limit=1, window=60)
        with patch("rag_api.time.monotonic", side_effect=[100.0, 101.0, 160.0]):
            self.assertEqual(limiter.allow(), (True, 0))
            self.assertEqual(limiter.allow(), (False, 60))
            self.assertEqual(limiter.allow(), (True, 0))

    def make_service(self):
        temporary = tempfile.TemporaryDirectory()
        root = Path(temporary.name)
        (root / "README.md").write_text(
            "O pedido aprovado muda para o status PAID.", encoding="utf-8"
        )
        embedder = FakeEmbedder()
        return temporary, RAGQueryService(LocalRAG(root, embedder), root), embedder

    def test_relevant_answer_preserves_structured_source(self):
        temporary, service, _ = self.make_service()
        self.addCleanup(temporary.cleanup)

        response = service.ask("Como funciona o pedido?")

        self.assertTrue(response["has_evidence"])
        self.assertEqual(response["documents_indexed"], 1)
        self.assertEqual(response["chunks_indexed"], 1)
        self.assertEqual(response["sources"][0]["file"], "README.md")
        self.assertEqual(response["sources"][0]["position"], 1)
        self.assertIn("similarity", response["sources"][0])
        self.assertIn("status PAID", response["sources"][0]["content"])

    def test_irrelevant_answer_returns_exact_absence_message(self):
        temporary, service, _ = self.make_service()
        self.addCleanup(temporary.cleanup)

        response = service.ask("Qual é a política de férias?")

        self.assertFalse(response["has_evidence"])
        self.assertEqual(response["answer"], NO_EVIDENCE)
        self.assertEqual(response["sources"], [])

    def test_empty_question_does_not_build_index_or_call_embedder(self):
        temporary, service, embedder = self.make_service()
        self.addCleanup(temporary.cleanup)

        response = service.ask("   ")

        self.assertEqual(response["answer"], NO_EVIDENCE)
        self.assertEqual(response["chunks_indexed"], 0)
        self.assertEqual(embedder.calls, 0)

    def test_question_length_is_limited_before_embedding(self):
        temporary, service, embedder = self.make_service()
        self.addCleanup(temporary.cleanup)

        with self.assertRaisesRegex(ValueError, "no máximo"):
            service.ask("x" * (MAX_QUESTION_CHARS + 1))

        self.assertEqual(embedder.calls, 0)

    def test_allowed_origins_are_trimmed_and_empty_values_ignored(self):
        origins = parse_allowed_origins(
            "http://localhost:5173, https://frontend.example.app, "
        )

        self.assertEqual(
            origins,
            frozenset({"http://localhost:5173", "https://frontend.example.app"}),
        )


class RAGHTTPTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        root = Path(self.temporary.name)
        (root / "README.md").write_text(
            "O pedido aprovado muda para o status PAID.", encoding="utf-8"
        )
        self.embedder = FakeEmbedder()
        service = RAGQueryService(LocalRAG(root, self.embedder), root)
        handler = make_handler(service, frozenset({"https://frontend.example.app"}))
        self.server = ThreadingHTTPServer(("127.0.0.1", 0), handler)
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()

    def tearDown(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=2)
        self.temporary.cleanup()

    def request(self, origin):
        body = json.dumps({"question": "Como funciona o pedido?"}).encode("utf-8")
        return Request(
            f"http://127.0.0.1:{self.server.server_port}/ask",
            data=body,
            method="POST",
            headers={"Content-Type": "application/json", "Origin": origin},
        )

    def test_allowed_origin_receives_evidence_and_cors_header(self):
        with urlopen(self.request("https://frontend.example.app"), timeout=2) as response:
            payload = json.load(response)

        self.assertTrue(payload["has_evidence"])
        self.assertEqual(
            response.headers["Access-Control-Allow-Origin"],
            "https://frontend.example.app",
        )

    def test_blocked_origin_returns_403_without_embedding_call(self):
        with self.assertRaises(HTTPError) as captured:
            urlopen(self.request("https://malicious.example"), timeout=2)

        error = captured.exception
        self.addCleanup(error.close)
        self.assertEqual(error.code, 403)
        self.assertEqual(self.embedder.calls, 0)

    def test_open_coach_failure_reports_only_stage(self):
        class FailingService:
            def ask(self, question):
                raise OpenCoachStepError("sanity_connect")

        self.server.RequestHandlerClass = make_handler(
            FailingService(), frozenset({"https://frontend.example.app"})
        )

        with self.assertRaises(HTTPError) as captured:
            urlopen(self.request("https://frontend.example.app"), timeout=2)

        error = captured.exception
        self.addCleanup(error.close)
        self.assertEqual(error.code, 503)
        self.assertEqual(
            json.load(error),
            {"error": "não foi possível consultar o OpenCoach", "stage": "sanity_connect"},
        )

    def test_nested_gemini_failure_reports_safe_provider_status(self):
        class FailingService:
            def ask(self, question):
                cause = ExceptionGroup(
                    "mcp", [GeminiRequestError(429, "RESOURCE_EXHAUSTED")]
                )
                raise OpenCoachStepError("gemini_selection", cause) from cause

        self.server.RequestHandlerClass = make_handler(
            FailingService(), frozenset({"https://frontend.example.app"})
        )

        with self.assertRaises(HTTPError) as captured:
            urlopen(self.request("https://frontend.example.app"), timeout=2)

        error = captured.exception
        self.addCleanup(error.close)
        self.assertEqual(error.code, 503)
        self.assertEqual(
            json.load(error),
            {
                "error": "não foi possível consultar o OpenCoach",
                "stage": "gemini_selection",
                "provider_http_status": 429,
                "provider_status": "RESOURCE_EXHAUSTED",
            },
        )

    def test_ask_rate_limit_returns_429_before_another_provider_call(self):
        class CountingService:
            calls = 0

            def ask(self, question):
                self.calls += 1
                return {"answer": question}

        service = CountingService()
        self.server.RequestHandlerClass = make_handler(
            service,
            frozenset({"https://frontend.example.app"}),
            AskRateLimiter(limit=1),
        )

        with urlopen(self.request("https://frontend.example.app"), timeout=2):
            pass
        with self.assertRaises(HTTPError) as captured:
            urlopen(self.request("https://frontend.example.app"), timeout=2)

        error = captured.exception
        self.addCleanup(error.close)
        self.assertEqual(error.code, 429)
        self.assertGreaterEqual(int(error.headers["Retry-After"]), 1)
        self.assertEqual(service.calls, 1)


if __name__ == "__main__":
    unittest.main()
