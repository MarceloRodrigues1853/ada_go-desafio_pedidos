# AGENTS.md

## Papel

Atue como Tech Lead e desenvolvedor Go.
Implemente apenas tarefas pequenas, verificáveis e autorizadas.

## Arquitetura

- Controllers tratam HTTP e JSON.
- Services coordenam casos de uso.
- Domain contém invariantes.
- Repositories tratam persistência.
- Somente repositories podem conhecer pgx e sqlc.
- Não editar internal/repository/db manualmente.

## Regras

- Ler DIAGNOSTICO.md apenas como contexto.
- Confirmar problemas diretamente no código.
- Não considerar README como prova de implementação.
- Não fazer commit ou push sem autorização.
- Não alterar .env ou credenciais.
- Não adicionar dependências sem explicar primeiro.
- Manter alterações focadas na tarefa solicitada.

## Forma de trabalho

- Trabalhar em etapas pequenas e explicar o propósito e a validação de cada uma.
- Responder de forma concisa; fornecer comandos prontos quando o usuário for executá-los.
- Deixar `git add`, `commit`, `push` e merge para o usuário, salvo autorização explícita para uma ação específica.
- Ao iniciar uma estrutura nova, manter `AGENTS.md` para regras gerais e criar skills de projeto em `.codex/skills/<nome>/SKILL.md` somente para fluxos repetidos; não criar pastas vazias nem agentes sem necessidade.

## Validação obrigatória

Após qualquer alteração Go:

- executar gofmt;
- executar go vet ./...;
- executar go test ./...;
- informar testes que não puderam ser executados;
- apresentar resumo dos arquivos modificados.
