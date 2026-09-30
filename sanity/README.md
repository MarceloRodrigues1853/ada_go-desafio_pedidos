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

## Teste de divergência histórica

`seed/historical-dlq-claim.ndjson` contém uma única afirmação marcada como
`outdated`: o diagnóstico anterior recomendava retry e DLQ, enquanto o código
atual envia mensagens rejeitadas sem requeue para a DLQ. A fonte histórica
referencia a afirmação atual `claim-dlq-current`; não deve ser tratada como
comportamento vigente.

Para adicionar apenas esse documento ao dataset já existente, execute na pasta
`sanity/` (com a sessão da CLI autenticada):

```bash
npx sanity datasets import seed/historical-dlq-claim.ndjson --dataset production
```

Não use `--replace` nem reimporte `initial-content.ndjson` neste passo. Depois,
na Knowledge Base do painel Sanity, escolha **Check for changes** para verificar
as fontes e examine **Issues**. A verificação não reescreve as entradas por si
só. Se as entradas ainda mostrarem a versão anterior e não houver uma issue
aplicável, use **Settings > Rebuild knowledge base**, sabendo que isso regenera
todas as entradas. Confira **Entries up to date** e a seção **Contexto histórico**
da entrada sobre DLQ antes de perguntar ao OpenCoach.

## Validação das citações da entrada sobre DLQ

Após reconstruir a Knowledge Base, confira as citações pelo cartão de origem;
os números podem mudar entre versões da entrada. A decisão arquitetural
`decision-dlq` documenta `messages_dlq_total` e o monitoramento das DLQs. A
afirmação `claim-dlq-current` confirma a existência da DLX, das filas e da
rejeição sem requeue, mas não descreve a métrica. A afirmação histórica
`claim-dlq-historical-diagnostic` descreve apenas o estado anterior.

Na revisão de 30/09/2026, a entrada gerada citou `claim-dlq-current` nas frases
sobre a métrica e o monitoramento, apesar de instruções de reescrita apontarem
para `decision-dlq`. A nota histórica citou a fonte histórica corretamente.
Portanto, a resposta do OpenCoach sobre a existência da DLQ foi validada, mas
a precisão dessas duas citações da entrada continua pendente. Não interprete
**Entries up to date** como aprovação da correspondência entre afirmação e
fonte. Evite repetir rebuilds sem uma mudança verificável na fonte ou no
processo de geração.
