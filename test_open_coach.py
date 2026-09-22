import unittest
from unittest.mock import patch

from open_coach import (
    GeminiCoach,
    KnowledgeEntry,
    NO_EVIDENCE,
    OpenCoachService,
    _knowledge_base_arguments,
    _tool_entries,
)


class FakeRetriever:
    def __init__(self, entries):
        self.entries = entries
        self.questions = []

    def retrieve(self, question):
        self.questions.append(question)
        return self.entries


class FakeGenerator:
    def __init__(self, response):
        self.response = response

    def answer(self, question, entries):
        return self.response


class OpenCoachServiceTests(unittest.TestCase):
    def test_gemini_string_false_does_not_claim_evidence(self):
        coach = GeminiCoach("test-key")
        with patch.object(
            coach,
            "_generate_json",
            return_value={
                "answer": "Inventada",
                "has_evidence": "false",
                "confidence": "high",
                "conflicts": None,
            },
        ):
            result = coach.answer("Pergunta", [KnowledgeEntry("regra/a", "Fonte")])

        self.assertFalse(result["has_evidence"])
        self.assertEqual(result["answer"], NO_EVIDENCE)
        self.assertEqual(result["confidence"], "low")
        self.assertEqual(result["conflicts"], [])

    def test_gemini_invalid_paths_are_rejected(self):
        coach = GeminiCoach("test-key")
        with patch.object(
            coach,
            "_generate_json",
            return_value={"kb_id": "kb123", "paths": "regra/a"},
        ):
            result = coach.select_entries("Pergunta", "Outline")

        self.assertEqual(result["paths"], [])

    def test_knowledge_base_read_uses_sanity_parameter_name(self):
        self.assertEqual(
            _knowledge_base_arguments({"kb_id": "kb123", "paths": ["saga/payment"]}),
            {"knowledgeBase": "kb123", "paths": ["saga/payment"]},
        )

    def test_answer_preserves_sanity_sources_and_conflicts(self):
        retriever = FakeRetriever(
            [KnowledgeEntry("saga/payment-failed", "O estoque é devolvido.")]
        )
        generator = FakeGenerator(
            {
                "answer": "A compensação devolve o estoque.",
                "has_evidence": True,
                "confidence": "high",
                "conflicts": ["Uma fonte antiga descreve a DLQ como pendente."],
            }
        )

        response = OpenCoachService(retriever, generator).ask(
            "Como funciona a compensação?"
        )

        self.assertTrue(response["has_evidence"])
        self.assertEqual(response["retrieval_mode"], "sanity-context")
        self.assertEqual(response["sources"][0]["file"], "saga/payment-failed")
        self.assertEqual(response["confidence"], "high")
        self.assertEqual(len(response["conflicts"]), 1)

    def test_missing_entries_refuses_without_calling_generator(self):
        retriever = FakeRetriever([])
        service = OpenCoachService(
            retriever,
            FakeGenerator(
                {
                    "answer": "não deveria ser usada",
                    "has_evidence": True,
                    "confidence": "high",
                    "conflicts": [],
                }
            ),
        )

        response = service.ask("Qual é a política de férias?")

        self.assertFalse(response["has_evidence"])
        self.assertEqual(response["answer"], NO_EVIDENCE)
        self.assertEqual(response["sources"], [])

    def test_empty_question_does_not_retrieve(self):
        retriever = FakeRetriever([])
        service = OpenCoachService(retriever, FakeGenerator({}))

        response = service.ask("   ")

        self.assertEqual(response["answer"], NO_EVIDENCE)
        self.assertEqual(retriever.questions, [])

    def test_tool_entries_falls_back_to_one_aggregate_when_blocks_do_not_match(self):
        class Block:
            text = "conteúdo agregado"

        class Result:
            content = [Block()]

        entries = _tool_entries(Result(), ["saga/create", "saga/payment"])

        self.assertEqual(len(entries), 1)
        self.assertEqual(entries[0].path, "saga/create, saga/payment")


if __name__ == "__main__":
    unittest.main()
