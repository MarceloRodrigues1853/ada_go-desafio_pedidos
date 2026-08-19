import os
import subprocess
from pathlib import Path
from datetime import datetime

from dotenv import load_dotenv
from crewai import Agent, Task, Crew, Process, LLM
from crewai.tools import tool


# ============================================================
# 0. CONFIGURAÇÃO
# ============================================================

load_dotenv()

GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY")

if not GOOGLE_API_KEY:
    raise RuntimeError(
        "GOOGLE_API_KEY não configurada no arquivo .env"
    )


# ============================================================
# 1. MODELO
# ============================================================

gemini_llm = LLM(
    model="gemini/gemini-3.1-flash-lite",
    api_key=GOOGLE_API_KEY,
    temperature=0.1,
)

print("🤖 Modelo: gemini/gemini-3.1-flash-lite")


# ============================================================
# 2. PROJETO
# ============================================================

PROJETO = Path.cwd().resolve()


# ============================================================
# 3. FERRAMENTA — ESTRUTURA
# ============================================================

@tool("listar_estrutura_projeto")
def listar_estrutura_projeto() -> str:
    """Lista a estrutura relevante do projeto."""

    ignorar = {
        ".git",
        "venv",
        ".venv",
        "__pycache__",
        "node_modules",
        ".idea",
        ".vscode",
    }

    arquivos = []

    for caminho in sorted(PROJETO.rglob("*")):

        if not caminho.is_file():
            continue

        if set(caminho.parts).intersection(ignorar):
            continue

        try:
            relativo = caminho.relative_to(PROJETO)
            arquivos.append(str(relativo))
        except ValueError:
            continue

    if not arquivos:
        return "Nenhum arquivo encontrado."

    limite = 600

    resultado = "\n".join(arquivos[:limite])

    if len(arquivos) > limite:
        resultado += (
            f"\n\n... {len(arquivos) - limite} arquivos adicionais."
        )

    return resultado


# ============================================================
# 4. FERRAMENTA — LER ARQUIVO
# ============================================================

@tool("ler_arquivo_projeto")
def ler_arquivo_projeto(caminho: str) -> str:
    """
    Lê um arquivo do projeto.
    Nunca permite acesso fora da raiz.
    """

    try:

        arquivo = (PROJETO / caminho).resolve()

        if not arquivo.is_relative_to(PROJETO):
            return "ERRO: acesso fora do projeto."

        if ".env" in arquivo.name.lower():
            return "ERRO: arquivo .env protegido."

        if ".git" in arquivo.parts:
            return "ERRO: arquivos .git protegidos."

        if not arquivo.exists():
            return f"Arquivo não encontrado: {caminho}"

        if not arquivo.is_file():
            return f"Não é um arquivo: {caminho}"

        tamanho_maximo = 50000

        tamanho = arquivo.stat().st_size

        if tamanho > tamanho_maximo:
            return (
                f"Arquivo muito grande: {caminho}\n"
                f"Tamanho: {tamanho} bytes\n"
                f"Limite: {tamanho_maximo} bytes"
            )

        conteudo = arquivo.read_text(
            encoding="utf-8",
            errors="ignore",
        )

        return (
            f"===== {caminho} =====\n\n"
            f"{conteudo}"
        )

    except Exception as erro:
        return f"Erro ao ler arquivo: {erro}"


# ============================================================
# 5. FERRAMENTA — GIT STATUS
# ============================================================

@tool("verificar_git_status")
def verificar_git_status() -> str:
    """Verifica branch e alterações do Git."""

    try:

        resultado = subprocess.run(
            [
                "git",
                "status",
                "--short",
                "--branch",
            ],
            cwd=PROJETO,
            capture_output=True,
            text=True,
            timeout=10,
        )

        if resultado.returncode != 0:
            return resultado.stderr

        return (
            resultado.stdout
            or "Repositório sem alterações pendentes."
        )

    except Exception as erro:
        return f"Erro no Git: {erro}"


# ============================================================
# 6. FERRAMENTA — TESTES
# ============================================================

@tool("executar_testes_go")
def executar_testes_go() -> str:
    """Executa go test ./..."""

    try:

        resultado = subprocess.run(
            ["go", "test", "./..."],
            cwd=PROJETO,
            capture_output=True,
            text=True,
            timeout=180,
        )

        saida = (
            resultado.stdout
            + "\n"
            + resultado.stderr
        )

        return (
            f"EXIT CODE: {resultado.returncode}\n\n"
            f"{saida[-15000:]}"
        )

    except subprocess.TimeoutExpired:
        return "ERRO: go test ./... excedeu 180 segundos."

    except Exception as erro:
        return f"Erro ao executar testes: {erro}"


# ============================================================
# 7. FERRAMENTA — GO VET
# ============================================================

@tool("executar_go_vet")
def executar_go_vet() -> str:
    """Executa go vet ./..."""

    try:

        resultado = subprocess.run(
            ["go", "vet", "./..."],
            cwd=PROJETO,
            capture_output=True,
            text=True,
            timeout=180,
        )

        saida = (
            resultado.stdout
            + "\n"
            + resultado.stderr
        )

        return (
            f"EXIT CODE: {resultado.returncode}\n\n"
            f"{saida[-15000:]}"
        )

    except subprocess.TimeoutExpired:
        return "ERRO: go vet ./... excedeu 180 segundos."

    except Exception as erro:
        return f"Erro ao executar go vet: {erro}"


# ============================================================
# 8. FERRAMENTA — LER DIAGNÓSTICO ANTERIOR
# ============================================================

@tool("ler_diagnostico_anterior")
def ler_diagnostico_anterior() -> str:
    """
    Lê o DIAGNOSTICO.md anterior para evitar
    repetir tarefas já concluídas.
    """

    arquivo = PROJETO / "DIAGNOSTICO.md"

    if not arquivo.exists():
        return "Nenhum diagnóstico anterior encontrado."

    try:

        tamanho_maximo = 60000

        if arquivo.stat().st_size > tamanho_maximo:
            return (
                "DIAGNOSTICO.md existe, mas é grande demais "
                "para leitura completa."
            )

        return arquivo.read_text(
            encoding="utf-8",
            errors="ignore",
        )

    except Exception as erro:
        return f"Erro ao ler diagnóstico anterior: {erro}"


# ============================================================
# 9. AGENTE
# ============================================================

agente_tutor_go = Agent(

    role="Tech Lead, Arquiteto e Auditor Go",

    goal="""
    Auditar tecnicamente o projeto Go e determinar
    qual é a próxima melhoria real que deve ser implementada.

    A decisão deve ser baseada exclusivamente em evidências
    encontradas no código, testes, infraestrutura e configuração.

    Nunca recomendar novamente uma funcionalidade que já esteja
    corretamente implementada.
    """,

    backstory="""
    Você é um Tech Lead experiente em:

    - Go
    - Clean Architecture
    - DDD
    - PostgreSQL
    - SQLC
    - RabbitMQ
    - Microsserviços
    - Saga
    - Idempotência
    - Observabilidade
    - Testes unitários e integração
    - Docker

    Você trabalha como um auditor técnico rigoroso.

    REGRAS OBRIGATÓRIAS:

    1. Nunca invente arquivos.
    2. Nunca invente funcionalidades.
    3. Nunca considere documentação como prova de implementação.
    4. Procure a implementação real no código.
    5. Diferencie:
       - IMPLEMENTADO
       - PARCIAL
       - NÃO IDENTIFICADO
    6. Consulte o diagnóstico anterior.
    7. Não repita tarefas já concluídas.
    8. Preserve a arquitetura existente.
    9. Não altere arquivos.
    10. Não implemente código.
    11. Execute os testes antes de concluir.
    12. Execute go vet antes de concluir.
    13. Escolha SOMENTE UMA próxima tarefa principal.
    14. A próxima tarefa deve ser pequena o suficiente para
        ser implementada em uma única etapa pelo OpenCode.
    15. Toda recomendação deve apontar os arquivos envolvidos.
    """,

    verbose=True,

    memory=False,

    allow_delegation=False,

    tools=[
        listar_estrutura_projeto,
        ler_arquivo_projeto,
        ler_diagnostico_anterior,
        verificar_git_status,
        executar_testes_go,
        executar_go_vet,
    ],

    llm=gemini_llm,

    max_iter=30,
)


# ============================================================
# 10. AUDITORIA
# ============================================================

data_hoje = datetime.now().strftime("%d/%m/%Y")

tarefa_auditoria = Task(

    description=f"""
    Faça uma auditoria completa e objetiva do projeto Go.

    Data: {data_hoje}

    ============================================================
    ETAPA 1 — CONTEXTO ANTERIOR
    ============================================================

    Primeiro leia o DIAGNOSTICO.md anterior.

    Use-o apenas como contexto.

    NÃO assuma que as tarefas descritas nele continuam
    pendentes.

    Verifique novamente no código.

    ============================================================
    ETAPA 2 — ESTRUTURA
    ============================================================

    Execute:

    1. listar_estrutura_projeto
    2. verificar_git_status

    Depois investigue os arquivos relevantes.

    ============================================================
    ETAPA 3 — ARQUIVOS IMPORTANTES
    ============================================================

    Analise, quando existirem:

    - go.mod
    - docker-compose.yml
    - README.md
    - cmd/app
    - cmd/payments
    - internal/domain
    - internal/service
    - internal/payments
    - internal/events
    - internal/repository
    - internal/infra
    - migrations
    - sqlc
    - testes

    Não leia arquivos aleatoriamente.

    Priorize arquivos relacionados às pendências
    encontradas no diagnóstico anterior.

    ============================================================
    ETAPA 4 — VALIDAR IMPLEMENTAÇÃO
    ============================================================

    Verifique especialmente:

    FASE 2
    - Payments realmente isolado?
    - Banco/schema?
    - cmd próprio?

    FASE 3
    - RabbitMQ?
    - publicação?
    - consumo?
    - idempotência?
    - tratamento de erros?

    FASE 4
    - Saga?
    - fluxo feliz?
    - fluxo de falha?
    - compensação?
    - estoque devolvido?
    - status do pedido?

    FASE 5
    - slog?
    - saga_id?
    - correlation?
    - logs nos serviços?
    - contexto?

    FASE 6
    - README?
    - arquitetura?
    - Mermaid?
    - comandos?
    - fluxo feliz?
    - fluxo de falha?

    ============================================================
    ETAPA 5 — EXECUTAR VALIDAÇÕES
    ============================================================

    Execute obrigatoriamente:

    go test ./...

    go vet ./...

    Se os testes falharem, investigue a causa.

    ============================================================
    ETAPA 6 — DECISÃO
    ============================================================

    Determine o estado real do projeto.

    Não tente encontrar uma pendência artificial.

    Se as fases principais estiverem implementadas,
    procure a próxima melhoria arquitetural REAL.

    Exemplos possíveis:

    - testes de integração reais
    - Docker Compose end-to-end
    - robustez de mensageria
    - retry
    - dead-letter queue
    - observabilidade
    - tratamento de falhas
    - transações
    - consistência
    - documentação
    - cobertura de testes
    - qualidade arquitetural

    Mas só recomende algo se houver evidência
    de que isso realmente é necessário.

    ============================================================
    FORMATO OBRIGATÓRIO
    ============================================================

    # Relatório de Auditoria de Arquitetura Go

    Data: {data_hoje}
    Auditor: Tech Lead & Arquiteto Go

    ## 1. Estado atual

    Resuma o estado real do projeto.

    ## 2. Evidências encontradas

    Liste arquivos e funcionalidades que comprovem
    suas conclusões.

    ## 3. Checklist das fases

    | Fase | Status | Evidência |
    |------|--------|-----------|

    Use somente:

    IMPLEMENTADO
    PARCIAL
    NÃO IDENTIFICADO

    ## 4. Validação

    Informe:

    - go test ./...
    - go vet ./...

    ## 5. Pendências reais

    Liste SOMENTE problemas ainda existentes.

    ## 6. Próximo passo recomendado

    Escolha UMA única tarefa principal.

    Explique:

    - objetivo
    - por que ela é necessária
    - arquivos envolvidos
    - comportamento esperado

    ## 7. Testes necessários

    Informe exatamente quais testes devem existir
    ou ser alterados.

    ## 8. Critério de conclusão

    Defina como saberemos que a tarefa terminou.

    ============================================================

    REGRA FINAL:

    O relatório será usado pelo OpenCode.

    Portanto, NÃO entregue uma lista gigante de melhorias.

    Entregue UMA próxima tarefa implementável.

    Não altere o código.
    """,

    expected_output="""
    Relatório técnico Markdown baseado em evidências reais
    encontradas no projeto.

    O relatório deve terminar com uma única tarefa
    principal recomendada para implementação.
    """,

    agent=agente_tutor_go,
)


# ============================================================
# 11. CREW
# ============================================================

equipe_tutora = Crew(

    agents=[agente_tutor_go],

    tasks=[tarefa_auditoria],

    process=Process.sequential,

    verbose=True,
)


# ============================================================
# 12. EXECUÇÃO
# ============================================================

if __name__ == "__main__":
    print("\n" + "=" * 60)
    print("🚀 INICIANDO AUDITORIA COM AGENTE TUTOR GO")
    print("=" * 60 + "\n")

    resultado = equipe_tutora.kickoff()

    caminho_relatorio = PROJETO / "DIAGNOSTICO.md"
    caminho_relatorio.write_text(
        str(resultado),
        encoding="utf-8"
    )

    print("\n" + "=" * 60)
    print(f"📊 Relatório salvo com sucesso em: {caminho_relatorio.name}")
    print("=" * 60 + "\n")

    print(resultado)