# Protótipo RAG do agente tutor

O protótipo adiciona recuperação sem LangChain ou LangGraph. Ele indexa somente
`README.md`, `DIAGNOSTICO.md` e arquivos `docs/**/*.md`, gera embeddings com o
modelo `gemini-embedding-001` e calcula similaridade de cosseno localmente.

## Como demonstrar

1. Ative o ambiente Python que já contém `crewai` e `python-dotenv`.
2. Configure `GOOGLE_API_KEY` no `.env` local. Não versione esse arquivo.
3. Execute `python agente_tutor.py`.
4. Durante a auditoria, procure no log a ferramenta `consultar_documentacao_rag`.
   A saída mostra cada trecho, seu arquivo de origem e a similaridade.
5. Para demonstrar ausência de evidência, consulte um assunto que não esteja na
   documentação; a ferramenta responde que não encontrou evidência suficiente.

O índice é recriado em memória a cada execução. Essa escolha mantém o protótipo
pequeno e evita persistir conteúdo sensível. A allowlist e a resolução de paths
impedem leitura de `.env`, `.git`, `AGENTS.md`, arquivos não Markdown, caminhos
fora da raiz e escapes por links simbólicos.

## Testes

Execute:

```powershell
python -m unittest -v test_rag_tutor.py
```

Os testes usam embeddings falsos e, portanto, não fazem chamadas externas nem
precisam de uma chave do Gemini.
