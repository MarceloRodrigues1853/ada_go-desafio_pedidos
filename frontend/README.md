# Frontend do Sistema de Pedidos

Painel operacional desenvolvido com React, TypeScript e Vite.

Na criação de pedidos, o painel permite simular pagamentos por cartão, Pix ou
boleto com aprovação ou recusa controlada. Nenhum dado bancário é coletado e
nenhuma cobrança real é realizada.

Ao criar o pedido, o resultado escolhido é processado automaticamente pelo
microsserviço Payments e pela Saga. O painel acompanha pedidos `PENDING` até a
transição para `PAID` ou `CANCELED`; não existem ações manuais de pagamento na
interface. As rotas manuais da API permanecem disponíveis por compatibilidade.

As telas de clientes e produtos oferecem busca local. A listagem de pedidos é
dividida em páginas de dez itens e atualizada automaticamente enquanto existir
algum pedido com status `PENDING`.

A tela **Tutor RAG** consulta a API documental separada. Ela exibe o arquivo de
origem, a posição e a similaridade de cada trecho recuperado, além do threshold
e dos totais indexados. Se nenhuma fonte atingir o limite, a interface informa
explicitamente a ausência de evidência. A recuperação documental orienta a
investigação; o comportamento real sempre deve ser confirmado no código.

## Configuração

```bash
cp .env.example .env
npm install
npm run dev
```

`VITE_API_URL` deve apontar para a API Go. O valor local padrão é
`http://localhost:8080`.

`VITE_RAG_API_URL` deve apontar para a API RAG. O valor local padrão é
`http://localhost:8081`. Na Vercel, configure essa variável com a URL pública do
serviço `desafio-pedidos-rag` no Render.

## Validação

```bash
npm run build
npm run lint
```
