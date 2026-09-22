# OpenCoach Knowledge Studio

Este Studio modela o conhecimento técnico consumido pelo OpenCoach. Ele não
armazena pedidos, clientes, produtos ou credenciais da aplicação.

- Projeto Sanity: `OpenCoach - Sistema de Pedidos`
- Project ID público: `ss56mini`
- Organização: `OpenCoach Lab` (`o6a7ecvjt`)
- Knowledge Base: `kbbKnMXYLs6b`
- Context MCP: `https://api.sanity.io/v1/context/organizations/o6a7ecvjt/mcp/open-coach`

1. Crie um projeto no Sanity e copie `.env.example` para `.env`.
2. Defina `SANITY_STUDIO_PROJECT_ID` e mantenha o dataset `production`.
3. Execute `npm install` e `npm run dev`.
4. Importe o conteúdo inicial confirmado no código:
   `npx sanity dataset import seed/initial-content.ndjson production --replace`.
5. Cadastre ou revise conteúdo e execute `npm run deploy-schema`.
6. No Dashboard, revise a Knowledge Base e reconstrua-a depois de mudanças no conteúdo.
7. Configure a URL do Context MCP acima e o token de organização somente no
   backend do OpenCoach.

Os schemas distinguem fatos atuais, desatualizados e disputados. As referências
para código devem usar caminhos do repositório e nunca segredos ou conteúdo de
arquivos `.env`.

O seed contém apenas fatos verificados diretamente no código. Os campos
`codeSources` preservam os caminhos e linhas usados como evidência para facilitar
revisões futuras; eles não concedem ao agente acesso de execução ao repositório.
