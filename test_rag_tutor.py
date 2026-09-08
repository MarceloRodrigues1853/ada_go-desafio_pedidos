import tempfile
import unittest
from io import StringIO
from pathlib import Path
from unittest.mock import patch

from demo_rag import load_api_key, run_demo
from rag_tutor import DEFAULT_MIN_SIMILARITY, NO_EVIDENCE, LocalRAG, cosine_similarity, discover_markdown_files, safe_markdown_path, split_markdown


class FakeEmbedder:
    def __init__(self):
        self.calls = 0

    def __call__(self, text, task_type):
        self.calls += 1
        lowered = text.lower()
        return [float("pedido" in lowered), float("coelho" in lowered), 0.1]


class ChunkTests(unittest.TestCase):
    def test_split_respects_size_and_overlap(self):
        chunks = split_markdown("um dois três quatro cinco seis", "README.md", chunk_size=16, overlap=5)
        self.assertGreater(len(chunks), 1)
        self.assertEqual([chunk.position for chunk in chunks], list(range(len(chunks))))
        self.assertTrue(all(chunk.source == "README.md" for chunk in chunks))
        self.assertIn(set(chunks[0].text.split()) & set(chunks[1].text.split()), ({"três"}, {"dois", "três"}))


class SearchTests(unittest.TestCase):
    def test_relevant_result_above_default_threshold(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("O pedido possui itens e status.", encoding="utf-8")
            (root / "DIAGNOSTICO.md").write_text("Um coelho aparece neste exemplo.", encoding="utf-8")
            rag = LocalRAG(root, FakeEmbedder())
            rag.build()
            results = rag.search("Como funciona o pedido?")
            self.assertEqual(results[0].chunk.source, "README.md")
            self.assertGreaterEqual(results[0].score, DEFAULT_MIN_SIMILARITY)
            self.assertGreater(cosine_similarity([1, 0], [1, 0]), cosine_similarity([1, 0], [0, 1]))

    def test_irrelevant_result_below_default_threshold_returns_exact_message(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("Pedidos", encoding="utf-8")
            rag = LocalRAG(root, FakeEmbedder())
            rag.build()
            self.assertEqual(rag.context("assunto desconhecido"), NO_EVIDENCE)

    def test_empty_query_returns_no_evidence_without_embedding_call(self):
        embedder = FakeEmbedder()
        rag = LocalRAG(Path.cwd(), embedder)

        self.assertEqual(rag.context("   "), NO_EVIDENCE)
        self.assertEqual(embedder.calls, 0)


class FileProtectionTests(unittest.TestCase):
    def test_discovers_only_allowlisted_markdown(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("ok", encoding="utf-8")
            (root / "AGENTS.md").write_text("segredo", encoding="utf-8")
            (root / ".env").write_text("KEY=secret", encoding="utf-8")
            (root / "docs").mkdir()
            (root / "docs" / "guia.md").write_text("guia", encoding="utf-8")
            found = {path.relative_to(root).as_posix() for path in discover_markdown_files(root)}
            self.assertEqual(found, {"README.md", "docs/guia.md"})

    def test_rejects_outside_git_env_and_non_allowlisted_files(self):
        with tempfile.TemporaryDirectory() as directory, tempfile.TemporaryDirectory() as outside:
            root = Path(directory)
            protected = [root / ".env", root / ".git" / "config.md", root / "AGENTS.md", Path(outside) / "fora.md"]
            for path in protected:
                with self.subTest(path=path), self.assertRaises(ValueError):
                    safe_markdown_path(root, path)


class DemoTests(unittest.TestCase):
    def test_demo_loads_only_api_key_from_dotenv(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / ".env").write_text(
                "OUTRA_VARIAVEL=nao_usar\nGOOGLE_API_KEY=chave-de-teste\n",
                encoding="utf-8",
            )
            with patch.dict("os.environ", {}, clear=True):
                self.assertEqual(load_api_key(root), "chave-de-teste")

    def test_demo_shows_counts_source_position_score_and_content(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("O pedido possui itens e status.", encoding="utf-8")
            output = StringIO()

            exit_code = run_demo(FakeEmbedder(), "Como funciona o pedido?", output, root, min_score=0.5)

            rendered = output.getvalue()
            self.assertEqual(exit_code, 0)
            self.assertIn("Documentos indexados: 1", rendered)
            self.assertIn("Trechos indexados: 1", rendered)
            self.assertIn("Arquivo: README.md", rendered)
            self.assertIn("Posição: 1", rendered)
            self.assertIn("Similaridade:", rendered)
            self.assertIn("O pedido possui itens e status.", rendered)

    def test_demo_reports_no_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("Pedidos", encoding="utf-8")
            output = StringIO()

            run_demo(FakeEmbedder(), "assunto desconhecido", output, root, min_score=0.99)

            self.assertIn(NO_EVIDENCE, output.getvalue())

    def test_demo_empty_question_does_not_call_embedder(self):
        output = StringIO()
        embedder = FakeEmbedder()

        exit_code = run_demo(embedder, "  ", output)

        self.assertEqual(exit_code, 0)
        self.assertEqual(embedder.calls, 0)
        self.assertIn("Pergunta vazia", output.getvalue())
        self.assertIn(NO_EVIDENCE, output.getvalue())


if __name__ == "__main__":
    unittest.main()
