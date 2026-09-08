import tempfile
import unittest
from pathlib import Path

from langgraph_rag import build_rag_graph
from rag_tutor import DEFAULT_MIN_SIMILARITY, LocalRAG, NO_EVIDENCE


class FakeEmbedder:
    def __init__(self):
        self.calls = 0

    def __call__(self, text, task_type):
        self.calls += 1
        lowered = text.lower()
        return [float("pedido" in lowered), float("férias" in lowered), 0.1]


class LangGraphRAGTests(unittest.TestCase):
    def make_index(self):
        temporary = tempfile.TemporaryDirectory()
        root = Path(temporary.name)
        (root / "README.md").write_text(
            "O pedido é pago e muda para o status PAID.", encoding="utf-8"
        )
        embedder = FakeEmbedder()
        rag = LocalRAG(root, embedder)
        chunk_count = rag.build()
        return temporary, rag, embedder, chunk_count

    def invoke(self, rag, question, chunk_count=1, threshold=DEFAULT_MIN_SIMILARITY):
        trace = []
        graph = build_rag_graph(rag, trace.append)
        state = graph.invoke(
            {
                "question": question,
                "threshold": threshold,
                "document_count": 1,
                "chunk_count": chunk_count,
            }
        )
        return graph, state, trace

    def test_valid_question_routes_to_format_results_and_preserves_source(self):
        temporary, rag, _, chunks = self.make_index()
        self.addCleanup(temporary.cleanup)

        _, state, trace = self.invoke(rag, "Como funciona o pedido?", chunks)

        self.assertTrue(state["has_evidence"])
        self.assertEqual(trace[-1], "formatar_resultados")
        self.assertIn("Fonte: README.md", state["answer"])
        self.assertIn("Posição: 1", state["answer"])
        self.assertIn("Similaridade:", state["answer"])
        self.assertIn("Conteúdo:", state["answer"])

    def test_irrelevant_question_routes_to_report_absence(self):
        temporary, rag, _, chunks = self.make_index()
        self.addCleanup(temporary.cleanup)

        _, state, trace = self.invoke(rag, "Qual é a política de férias?", chunks)

        self.assertFalse(state["has_evidence"])
        self.assertEqual(trace[-1], "informar_ausencia")
        self.assertEqual(state["answer"], NO_EVIDENCE)

    def test_empty_question_does_not_call_embedder(self):
        embedder = FakeEmbedder()
        rag = LocalRAG(Path.cwd(), embedder)

        _, state, trace = self.invoke(rag, "   ", chunk_count=0)

        self.assertEqual(embedder.calls, 0)
        self.assertEqual(state["answer"], NO_EVIDENCE)
        self.assertEqual(trace[-1], "informar_ausencia")

    def test_custom_threshold_is_respected(self):
        temporary, rag, _, chunks = self.make_index()
        self.addCleanup(temporary.cleanup)

        _, state, trace = self.invoke(rag, "pedido", chunks, threshold=1.01)

        self.assertEqual(state["threshold"], 1.01)
        self.assertFalse(state["has_evidence"])
        self.assertEqual(trace[-1], "informar_ausencia")

    def test_recovery_error_is_recorded_in_state(self):
        rag = LocalRAG(Path.cwd(), FakeEmbedder())

        def fail_search(*args, **kwargs):
            raise RuntimeError("erro simulado")

        rag.search = fail_search
        _, state, trace = self.invoke(rag, "pedido")

        self.assertIn("erro simulado", state["error"])
        self.assertEqual(state["answer"], NO_EVIDENCE)
        self.assertEqual(trace[-1], "informar_ausencia")

    def test_flow_reaches_end_after_conditional_branch(self):
        temporary, rag, _, chunks = self.make_index()
        self.addCleanup(temporary.cleanup)

        graph, state, trace = self.invoke(rag, "pedido", chunks)
        edges = {(edge.source, edge.target) for edge in graph.get_graph().edges}

        self.assertIn(("formatar_resultados", "__end__"), edges)
        self.assertIn(("informar_ausencia", "__end__"), edges)
        self.assertEqual(
            trace,
            [
                "validar_pergunta",
                "recuperar_contexto",
                "verificar_evidencia",
                "formatar_resultados",
            ],
        )
        self.assertTrue(state["answer"])


if __name__ == "__main__":
    unittest.main()
