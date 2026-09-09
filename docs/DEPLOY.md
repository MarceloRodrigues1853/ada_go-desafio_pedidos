# Deploy do protótipo

Esta configuração mantém a arquitetura atual sem trocar banco, broker ou regras
de negócio:

- frontend React/Vite na Vercel;
- API e consumidor Payments como dois Web Services no Render;
- PostgreSQL no Neon;
- RabbitMQ no CloudAMQP.

O TiDB não faz parte desta etapa. Apesar de oferecer protocolo compatível com
MySQL, o projeto usa PostgreSQL, pgx, sqlc e migrations PostgreSQL. Adotá-lo
exigiria uma migração de persistência que não é necessária para demonstrar a
Saga.

## 1. Banco PostgreSQL no Neon

Crie um projeto gratuito e copie a connection string PostgreSQL com SSL. Antes
de publicar os serviços, aplique as migrations a partir do Git Bash. Não grave
a URL no repositório nem no histórico do shell:

```bash
read -s -p "DB_URL do Neon: " DB_URL
echo
export DB_URL
migrate -path migrations -database "$DB_URL" -verbose up
unset DB_URL
```

O plano gratuito e seus limites podem mudar. Consulte
<https://neon.com/pricing> antes da demonstração.

## 2. RabbitMQ no CloudAMQP

Crie uma instância Little Lemur e copie a URL AMQP. Ela será informada como
`RABBITMQ_URL` nos dois serviços do Render. Nunca adicione essa URL ao `.env`
versionado, ao `render.yaml` ou a capturas de tela.

Limites atuais do plano estão em <https://www.cloudamqp.com/plans.html>.

## 3. API e Payments no Render

Crie um Blueprint apontando para este repositório. O `render.yaml` define:

- `desafio-pedidos-api`, construído com `Dockerfile`;
- `desafio-pedidos-payments`, construído com `Dockerfile.payments`;
- região `ohio`, alinhada ao Neon e ao CloudAMQP;
- health check `/health` nos dois serviços;
- credenciais solicitadas no painel por meio de `sync: false`.

No primeiro deploy, informe:

| Serviço | Variável | Valor |
| --- | --- | --- |
| API | `DB_URL` | URL PostgreSQL do Neon |
| API | `RABBITMQ_URL` | URL AMQP do CloudAMQP |
| API | `CORS_ALLOWED_ORIGINS` | URL HTTPS do frontend na Vercel |
| Payments | `DB_URL` | A mesma URL PostgreSQL |
| Payments | `RABBITMQ_URL` | A mesma URL AMQP |

Não configure `PORT`: o Render fornece essa variável. No Payments ela tem
precedência sobre `METRICS_PORT`.

O plano gratuito do Render entra em suspensão após inatividade e possui uma
franquia mensal compartilhada. Um Web Service que executa um consumidor não é
uma substituição de produção para um Background Worker. Antes da demonstração,
abra primeiro `https://SEU-PAYMENTS.onrender.com/health` e depois
`https://SUA-API.onrender.com/health`; espere ambos responderem antes de criar
o pedido. O processamento pode demorar durante o cold start.

Consulte <https://render.com/docs/free> e
<https://render.com/docs/blueprint-spec> para os limites e campos atuais.

## 4. Frontend na Vercel

Importe o mesmo repositório e configure:

- Root Directory: `frontend`;
- Framework Preset: Vite;
- Build Command: `npm run build`;
- Output Directory: `dist`;
- variável `VITE_API_URL`: URL HTTPS pública da API no Render, sem barra final.

Depois que a Vercel informar a URL pública, volte ao serviço da API no Render e
confirme que `CORS_ALLOWED_ORIGINS` contém exatamente essa origem. Faça novo
deploy do frontend se alterar `VITE_API_URL`, pois variáveis `VITE_` são
incorporadas durante o build.

Referência: <https://vercel.com/docs/frameworks/frontend/vite>.

## Checklist da demonstração

1. Confirmar que as migrations foram aplicadas no Neon.
2. Abrir `/health` do Payments e aguardar `{"status":"ok"}`.
3. Abrir `/health` da API e aguardar resposta saudável.
4. Abrir o frontend e confirmar `API online`.
5. Criar cliente e produto.
6. Criar um pedido aprovado e observar `PENDING` mudar para `PAID`.
7. Criar um pedido recusado e observar `PENDING` mudar para `CANCELED`.
8. Confirmar que somente o pedido aprovado reduz o estoque definitivamente.

Essa topologia é adequada para portfólio e demonstração, não para produção:
serviços gratuitos podem suspender, têm cotas, não oferecem as mesmas garantias
de disponibilidade e podem mudar de condições.
