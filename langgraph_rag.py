"""Orquestra o RAG existente com um grafo de estados explícito."""

from __future__ import annotations

from typing import Callable, TypedDict

from langgraph.graph import END, START, StateGraph

from rag_tutor import DEFAULT_MIN_SIMILARITY, LocalRAG, NO_EVIDENCE, SearchResult


class RAGGraphState(TypedDict, total=False):
    question: str
    threshold: float
    document_count: int
    chunk_count: int
    results: list[SearchResult]
    has_evidence: bool
    answer: str
    error: str | None


NodeObserver = Callable[[str], None]


class RAGGraphNodes:
    """Nós pequenos que compõem um LocalRAG sem alterar seu comportamento."""

    def __init__(self, rag: LocalRAG, observer: NodeObserver | None = None):
        self.rag = rag
        self.observer = observer

    def _observed(self, name: str) -> None:
        if self.observer:
            self.observer(name)

    def validar_pergunta(self, state: RAGGraphState) -> dict:
        self._observed("validar_pergunta")
        return {
            "question": state.get("question", "").strip(),
            "threshold": state.get("threshold", DEFAULT_MIN_SIMILARITY),
            "error": None,
        }

    def recuperar_contexto(self, state: RAGGraphState) -> dict:
        self._observed("recuperar_contexto")
        if not state["question"]:
            return {"results": []}
        try:
            results = self.rag.search(
                state["question"],
                limit=3,
                min_score=state["threshold"],
            )
            return {"results": results}
        except Exception as error:
            return {"results": [], "error": f"Falha controlada na recuperação: {error}"}

    def verificar_evidencia(self, state: RAGGraphState) -> dict:
        self._observed("verificar_evidencia")
        has_evidence = not state.get("error") and any(
            result.score >= state["threshold"] for result in state.get("results", [])
        )
        return {"has_evidence": bool(has_evidence)}

    def formatar_resultados(self, state: RAGGraphState) -> dict:
        self._observed("formatar_resultados")
        blocks = []
        for result in state.get("results", []):
            blocks.append(
                f"Fonte: {result.chunk.source}\n"
                f"Posição: {result.chunk.position + 1}\n"
                f"Similaridade: {result.score:.3f}\n"
                f"Conteúdo: {result.chunk.text}"
            )
        return {"answer": "\n\n---\n\n".join(blocks)}

    def informar_ausencia(self, state: RAGGraphState) -> dict:
        self._observed("informar_ausencia")
        return {"answer": NO_EVIDENCE}


def escolher_saida(state: RAGGraphState) -> str:
    return "com_evidencia" if state.get("has_evidence", False) else "sem_evidencia"


def build_rag_graph(rag: LocalRAG, observer: NodeObserver | None = None):
    """Compila o fluxo LangGraph usando o LocalRAG fornecido."""
    nodes = RAGGraphNodes(rag, observer)
    graph = StateGraph(RAGGraphState)
    graph.add_node("validar_pergunta", nodes.validar_pergunta)
    graph.add_node("recuperar_contexto", nodes.recuperar_contexto)
    graph.add_node("verificar_evidencia", nodes.verificar_evidencia)
    graph.add_node("formatar_resultados", nodes.formatar_resultados)
    graph.add_node("informar_ausencia", nodes.informar_ausencia)

    graph.add_edge(START, "validar_pergunta")
    graph.add_edge("validar_pergunta", "recuperar_contexto")
    graph.add_edge("recuperar_contexto", "verificar_evidencia")
    graph.add_conditional_edges(
        "verificar_evidencia",
        escolher_saida,
        {
            "com_evidencia": "formatar_resultados",
            "sem_evidencia": "informar_ausencia",
        },
    )
    graph.add_edge("formatar_resultados", END)
    graph.add_edge("informar_ausencia", END)
    return graph.compile()
