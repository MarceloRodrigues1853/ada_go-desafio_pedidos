# Frontend do Sistema de Pedidos

Painel operacional desenvolvido com React, TypeScript e Vite.

Na criação de pedidos, o painel permite simular pagamentos por cartão, Pix ou
boleto com aprovação ou recusa controlada. Nenhum dado bancário é coletado e
nenhuma cobrança real é realizada.

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
