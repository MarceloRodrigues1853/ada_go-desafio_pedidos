import os
from pathlib import Path
from dotenv import load_dotenv
from crewai import Agent, Task, Crew, Process, LLM
from crewai.tools import tool

# ============================================================
# 0. CONFIGURAÇÃO
# ============================================================

load_dotenv()

GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY")
if not GOOGLE_API_KEY:
    raise RuntimeError("GOOGLE_API_KEY não configurada no arquivo .env")

# ============================================================
# 1. MODELO
# ============================================================

# Utilizando identificador válido da API Gemini
gemini_llm = LLM(
    model="gemini/gemini-3.1-flash-lite",
    api_key=GOOGLE_API_KEY,
    temperature=0.1,
)

print("🤖 Modelo selecionado: gemini/gemini-3.1-flash-lite")

# ============================================================
# 2. DIRETÓRIO DO PROJETO
# ============================================================

PROJETO = Path.cwd().resolve()

# ============================================================
# 3. FERRAMENTA — ESTRUTURA DO PROJETO
# ============================================================

@tool("listar_estrutura_projeto")
def listar_estrutura_projeto() -> str:
    """
    Lista a estrutura de arquivos e diretórios do projeto atual em modo somente leitura.
    """
    ignorar = {
        ".git", "venv", ".venv", "__pycache__", 
        "node_modules", ".idea", ".vscode"
    }

    resultado = []
    for caminho in sorted(PROJETO.rglob("*")):
        if set(caminho.parts).intersection(ignorar):
            continue
        try:
            relativo = caminho.relative_to(PROJETO)
            resultado.append(str(relativo))
        except ValueError:
            continue

    if not resultado:
        return "Nenhum arquivo encontrado."

    limite = 500
    texto = "\n".join(resultado[:limite])
    if len(resultado) > limite:
        texto += f"\n\n... {len(resultado) - limite} itens adicionais não exibidos."

    return texto

# ============================================================
# 4. FERRAMENTA — LER ARQUIVO
# ============================================================

@tool("ler_arquivo_projeto")
def ler_arquivo_projeto(caminho: str) -> str:
    """
    Lê o conteúdo de um arquivo do projeto em modo somente leitura.
    Parâmetro: caminho (str) relativo à raiz do projeto.
    """
    try:
        arquivo = (PROJETO / caminho).resolve()

        # Validação robusta de escopo de diretório
        if not arquivo.is_relative_to(PROJETO):
            return "ERRO: acesso fora do diretório do projeto não permitido."

        # Proteções de segurança
        if ".env" in arquivo.name.lower():
            return "ERRO: arquivos .env não podem ser lidos."

        if ".git" in arquivo.parts:
            return "ERRO: arquivos dentro de .git não podem ser lidos."

        if not arquivo.exists():
            return f"Arquivo não encontrado: {caminho}"

        if not arquivo.is_file():
            return f"O caminho não é um arquivo: {caminho}"

        # Limite de tamanho por arquivo (35 KB)
        tamanho_maximo = 35000
        if arquivo.stat().st_size > tamanho_maximo:
            return (
                f"Arquivo muito grande para leitura completa: {caminho} "
                f"({arquivo.stat().st_size} bytes, limite: {tamanho_maximo} bytes)"
            )

        conteudo = arquivo.read_text(encoding="utf-8", errors="ignore")
        return f"===== {caminho} =====\n\n{conteudo}"

    except Exception as erro:
        return f"Erro ao ler {caminho}: {erro}"

# ============================================================
# 5. FERRAMENTA — GIT STATUS
# ============================================================

@tool("verificar_git_status")
def verificar_git_status() -> str:
    """
    Verifica o estado atual do repositório Git (arquivos modificados e branch).
    """
    import subprocess

    try:
        resultado = subprocess.run(
            ["git", "status", "--short", "--branch"],
            cwd=PROJETO,
            capture_output=True,
            text=True,
            timeout=10,
        )
        if resultado.returncode != 0:
            return f"Erro ao executar git status:\n{resultado.stderr}"
        return resultado.stdout or "Repositório sem alterações pendentes."
    except Exception as erro:
        return f"Erro ao verificar Git: {erro}"

# ============================================================
# 6. AGENTE TUTOR
# ============================================================

agente_tutor_go = Agent(
    role="Tech Lead e Auditor de Arquitetura Go",
    goal="""
    Auditar a base de código do projeto Go, verificar a implementação 
    real de cada fase arquitetural e especificar com precisão a próxima tarefa.
    """,
    backstory="""
    Você é um Tech Lead especialista em Go, DDD, Clean Architecture, 
    microsserviços e sistemas distribuídos. Seu papel é atuar como auditor estrito:
    você só afirma que algo existe se encontrar o código correspondente.
    """,
    verbose=True,
    memory=False,
    allow_delegation=False,
    tools=[
        listar_estrutura_projeto,
        ler_arquivo_projeto,
        verificar_git_status,
    ],
    llm=gemini_llm,
    max_iter=25,
)

# ============================================================
# 7. TAREFA
# ============================================================

tarefa_auditoria = Task(
    description="""
    Execute uma auditoria completa no projeto Go:
    1. Liste a estrutura de arquivos com `listar_estrutura_projeto`.
    2. Verifique o status do repositório com `verificar_git_status`.
    3. Leia os arquivos-chave (`go.mod`, `docker-compose.yml`, `internal/...`, etc.).
    4. Avalie o estado das seguintes fases:
       - FASE 2: Microsserviço Payments (separação de serviço, cmd próprio, schema isolado)
       - FASE 3: Mensageria (Broker, eventos publicados/consumidos, idempotência)
       - FASE 4: Saga (Orquestração/Coreografia, fluxos de sucesso e compensação)
       - FASE 5: Logs Estruturados (slog com correlation/saga_id)
       - FASE 6: Documentação e Demonstrações
    5. Estruture o relatório com:
       - Fase atual concluída
       - Checklist detalhado (IMPLEMENTADO, PARCIAL, NÃO IDENTIFICADO)
       - O que realmente falta
       - **Próximo passo de implementação** detalhado com arquivos e testes a criar.
    """,
    expected_output="Relatório técnico em Markdown com diagnóstico e plano de ação único.",
    agent=agente_tutor_go,
)

equipe_tutora = Crew(
    agents=[agente_tutor_go],
    tasks=[tarefa_auditoria],
    process=Process.sequential,
    verbose=True,
)

# ============================================================
# 8. EXECUÇÃO E PERSISTÊNCIA
# ============================================================

if __name__ == "__main__":
    print("\n" + "=" * 60)
    print("🚀 INICIANDO AUDITORIA COM AGENTE TUTOR GO")
    print("=" * 60 + "\n")

    resultado = equipe_tutora.kickoff()

    # Salva o resultado em arquivo para consumo direto pelo OpenCode
    caminho_relatorio = PROJETO / "DIAGNOSTICO.md"
    caminho_relatorio.write_text(str(resultado), encoding="utf-8")

    print("\n" + "=" * 60)
    print(f"📊 Relatório salvo com sucesso em: {caminho_relatorio.name}")
    print("=" * 60 + "\n")
    print(resultado)

resultado = equipe_tutora.kickoff()

# Salva para o OpenCode ler diretamente
with open("DIAGNOSTICO.md", "w", encoding="utf-8") as f:
    f.write(str(resultado))

print("✅ DIAGNOSTICO.md gerado com sucesso!")