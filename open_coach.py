"""OpenCoach fundamentado em uma Knowledge Base servida pelo Sanity Context."""

from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Protocol
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


NO_EVIDENCE = "Não encontrei evidência suficiente na base de conhecimento."


class OpenCoachStepError(RuntimeError):
    def __init__(self, stage: str, cause: BaseException | None = None):
        super().__init__(f"Falha na etapa {stage}")
        self.stage = stage
        gemini_error = _find_gemini_error(cause)
        self.provider_http_status = gemini_error.http_status if gemini_error else None
        self.provider_status = gemini_error.provider_status if gemini_error else None


class GeminiRequestError(RuntimeError):
    def __init__(self, http_status: int, provider_status: str | None):
        super().__init__(f"Gemini retornou HTTP {http_status}")
        self.http_status = http_status
        self.provider_status = provider_status


def _find_gemini_error(error: BaseException | None) -> GeminiRequestError | None:
    pending = [error] if error is not None else []
    seen: set[int] = set()
    while pending:
        current = pending.pop()
        if id(current) in seen:
            continue
        seen.add(id(current))
        if isinstance(current, GeminiRequestError):
            return current
        if isinstance(current, BaseExceptionGroup):
            pending.extend(current.exceptions)
        if current.__cause__ is not None:
            pending.append(current.__cause__)
        if current.__context__ is not None:
            pending.append(current.__context__)
    return None


@dataclass(frozen=True)
class KnowledgeEntry:
    path: str
    content: str


class ContextRetriever(Protocol):
    def retrieve(self, question: str) -> list[KnowledgeEntry]: ...


class AnswerGenerator(Protocol):
    def answer(self, question: str, entries: list[KnowledgeEntry]) -> dict[str, Any]: ...


class GeminiCoach:
    """Gera a resposta final sem enviar credenciais ao navegador."""

    def __init__(self, api_key: str, model: str = "gemini-3.1-flash-lite", timeout: int = 60):
        if not api_key:
            raise ValueError("GOOGLE_API_KEY não configurada")
        self.api_key = api_key
        self.model = model
        self.timeout = timeout

    def answer(self, question: str, entries: list[KnowledgeEntry]) -> dict[str, Any]:
        evidence = "\n\n".join(
            f"SOURCE_PATH: {entry.path}\n{entry.content}" for entry in entries
        )
        prompt = f"""Você é o OpenCoach, tutor técnico do Sistema de Pedidos.
Responda somente com base nas evidências abaixo. Não complete lacunas com conhecimento geral.
Se as fontes divergirem, descreva a divergência em conflicts. Se não sustentarem a resposta,
retorne has_evidence=false. Produza JSON válido com as chaves answer (string),
has_evidence (boolean), confidence (low|medium|high) e conflicts (array de strings).

PERGUNTA:
{question}

EVIDÊNCIAS:
{evidence}
"""
        payload = self._generate_json(prompt)
        answer = payload.get("answer")
        has_evidence = (
            payload.get("has_evidence") is True
            and isinstance(answer, str)
            and bool(answer.strip())
        )
        conflicts = payload.get("conflicts")
        confidence = payload.get("confidence")
        return {
            "answer": answer.strip() if has_evidence else NO_EVIDENCE,
            "has_evidence": has_evidence,
            "confidence": (
                confidence
                if has_evidence and isinstance(confidence, str) and confidence in {"low", "medium", "high"}
                else "low"
            ),
            "conflicts": (
                [item for item in conflicts if isinstance(item, str)]
                if isinstance(conflicts, list)
                else []
            ),
        }

    def select_entries(self, question: str, outline: str, limit: int = 8) -> dict[str, Any]:
        prompt = f"""Selecione as entradas mais úteis para responder à pergunta usando o outline
da Knowledge Base. Copie o Knowledge Base ID e os paths exatamente como aparecem.
Retorne somente JSON com kb_id (string) e paths (array com no máximo {limit} strings).
Não invente paths.

PERGUNTA:
{question}

OUTLINE:
{outline}
"""
        payload = self._generate_json(prompt)
        raw_paths = payload.get("paths")
        paths = (
            list(dict.fromkeys(path for path in raw_paths if isinstance(path, str) and path.strip()))
            if isinstance(raw_paths, list)
            else []
        )
        kb_id = payload.get("kb_id")
        return {
            "kb_id": kb_id if isinstance(kb_id, str) else "",
            "paths": paths[:limit],
        }

    def _generate_json(self, prompt: str) -> dict[str, Any]:
        url = (
            "https://generativelanguage.googleapis.com/v1beta/models/"
            f"{self.model}:generateContent"
        )
        body = json.dumps(
            {
                "contents": [{"parts": [{"text": prompt}]}],
                "generationConfig": {"responseMimeType": "application/json"},
            }
        ).encode("utf-8")
        request = Request(
            url,
            data=body,
            method="POST",
            headers={"Content-Type": "application/json", "x-goog-api-key": self.api_key},
        )
        try:
            with urlopen(request, timeout=self.timeout) as response:
                payload = json.load(response)
            text = payload["candidates"][0]["content"]["parts"][0]["text"]
            decoded = json.loads(text)
            if not isinstance(decoded, dict):
                raise ValueError("resposta JSON não é um objeto")
            return decoded
        except HTTPError as error:
            detail = error.read(8192).decode("utf-8", errors="replace")
            try:
                provider_error = json.loads(detail).get("error", {})
                provider_status = provider_error.get("status") if isinstance(provider_error, dict) else None
            except (ValueError, AttributeError):
                provider_status = None
            if not isinstance(provider_status, str) or not provider_status.replace("_", "").isupper():
                provider_status = None
            raise GeminiRequestError(error.code, provider_status) from error
        except URLError as error:
            raise RuntimeError(f"falha ao acessar Gemini: {error.reason}") from error
        except (KeyError, IndexError, TypeError, ValueError, json.JSONDecodeError) as error:
            raise RuntimeError("resposta inválida do Gemini") from error


class SanityContextRetriever:
    """Consulta as ferramentas de Knowledge Base do Sanity Context por MCP."""

    def __init__(self, endpoint: str, token: str, router: GeminiCoach):
        if not endpoint or not token:
            raise ValueError("endpoint e token do Sanity Context são obrigatórios")
        self.endpoint = endpoint
        self.token = token
        self.router = router

    def retrieve(self, question: str) -> list[KnowledgeEntry]:
        import asyncio

        return asyncio.run(self._retrieve(question))

    async def _retrieve(self, question: str) -> list[KnowledgeEntry]:
        stage = "sanity_connect"
        try:
            import httpx2
            from mcp import Client
            from mcp.client.streamable_http import streamable_http_client
        except ImportError as error:
            raise OpenCoachStepError("mcp_dependency") from error

        try:
            async with httpx2.AsyncClient(
                headers={"Authorization": f"Bearer {self.token}"},
                timeout=httpx2.Timeout(30.0, read=300.0),
            ) as http_client:
                transport = streamable_http_client(self.endpoint, http_client=http_client)
                async with Client(transport) as client:
                    stage = "sanity_outline"
                    outline_result = await client.call_tool("initial_context", {})
                    if outline_result.is_error:
                        raise RuntimeError("Sanity Context recusou initial_context")
                    outline = _tool_text(outline_result)
                    stage = "gemini_selection"
                    selection = self.router.select_entries(question, outline)
                    if not selection["kb_id"] or not selection["paths"]:
                        return []
                    stage = "sanity_read"
                    result = await client.call_tool(
                        "knowledge_base_read",
                        _knowledge_base_arguments(selection),
                    )
                    if result.is_error:
                        raise RuntimeError("Sanity Context recusou knowledge_base_read")
                    return _tool_entries(result, selection["paths"])
        except Exception as error:
            raise OpenCoachStepError(stage, error) from error


def _knowledge_base_arguments(selection: dict[str, Any]) -> dict[str, Any]:
    return {"knowledgeBase": selection["kb_id"], "paths": selection["paths"]}


def _tool_text(result: Any) -> str:
    blocks = getattr(result, "content", [])
    texts = [str(block.text) for block in blocks if getattr(block, "text", None)]
    if not texts:
        raise RuntimeError("Sanity Context não retornou conteúdo textual")
    return "\n\n".join(texts)


def _tool_entries(result: Any, paths: list[str]) -> list[KnowledgeEntry]:
    blocks = getattr(result, "content", [])
    texts = [str(block.text) for block in blocks if getattr(block, "text", None)]
    if not texts:
        raise RuntimeError("Sanity Context não retornou conteúdo textual")
    if len(texts) == len(paths):
        return [KnowledgeEntry(path=path, content=text) for path, text in zip(paths, texts)]
    return [KnowledgeEntry(path=", ".join(paths), content="\n\n".join(texts))]


class OpenCoachService:
    def __init__(self, retriever: ContextRetriever, generator: AnswerGenerator):
        self.retriever = retriever
        self.generator = generator

    def ask(self, question: str) -> dict[str, Any]:
        normalized = question.strip()
        if not normalized:
            return self._empty()
        if len(normalized) > 500:
            raise ValueError("a pergunta deve ter no máximo 500 caracteres")

        entries = self.retriever.retrieve(normalized)
        if not entries:
            return self._empty()

        try:
            generated = self.generator.answer(normalized, entries)
        except Exception as error:
            raise OpenCoachStepError("gemini_answer", error) from error
        has_evidence = bool(generated["has_evidence"])
        return {
            "answer": generated["answer"] if has_evidence else NO_EVIDENCE,
            "has_evidence": has_evidence,
            "confidence": generated["confidence"],
            "conflicts": generated["conflicts"],
            "sources": [
                {"file": entry.path, "position": index + 1, "content": entry.content}
                for index, entry in enumerate(entries)
            ],
            "retrieval_mode": "sanity-context",
        }

    @staticmethod
    def _empty() -> dict[str, Any]:
        return {
            "answer": NO_EVIDENCE,
            "has_evidence": False,
            "confidence": "low",
            "conflicts": [],
            "sources": [],
            "retrieval_mode": "sanity-context",
        }
