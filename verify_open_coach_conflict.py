"""Verificação manual de conflitos com fontes fictícias, sem alterar o Sanity."""

import json
import os

from open_coach import GeminiCoach, KnowledgeEntry


def main() -> int:
    api_key = os.getenv("GOOGLE_API_KEY", "")
    if not api_key:
        print("GOOGLE_API_KEY ausente neste terminal.")
        return 2

    entries = [
        KnowledgeEntry(
            "teste-ficticio/politica-a",
            "Documento A: neste cenário fictício, o pedido pode ser cancelado "
            "até 10 dias após a compra.",
        ),
        KnowledgeEntry(
            "teste-ficticio/politica-b",
            "Documento B: neste mesmo cenário fictício, o pedido pode ser "
            "cancelado somente até 2 dias após a compra.",
        ),
    ]
    question = "Qual é o prazo de cancelamento do pedido neste cenário fictício?"
    result = GeminiCoach(api_key).answer(question, entries)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    if not result["conflicts"]:
        print("FALHA: o conflito entre os dois documentos não foi apontado.")
        return 1
    print("OK: o conflito entre os dois documentos foi apontado.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
