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

## Configuração

```bash
cp .env.example .env
npm install
npm run dev
```

`VITE_API_URL` deve apontar para a API Go. O valor local padrão é
`http://localhost:8080`.

## Validação

```bash
npm run build
npm run lint
```
