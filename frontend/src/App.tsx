import { type FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { api, type ApiStatus, type Client, type Order, type Product } from './api'
import './App.css'

type View = 'dashboard' | 'clientes' | 'produtos' | 'pedidos'

const navigation: { id: View; label: string; marker: string }[] = [
  { id: 'dashboard', label: 'Visão geral', marker: '01' },
  { id: 'clientes', label: 'Clientes', marker: '02' },
  { id: 'produtos', label: 'Produtos', marker: '03' },
  { id: 'pedidos', label: 'Pedidos', marker: '04' },
]

function formatDate(value: string) {
  return new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value))
}

function formatMoney(value: number) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value)
}

function shortId(value: string) {
  return value.length > 13 ? `${value.slice(0, 8)}…` : value
}

function App() {
  const [view, setView] = useState<View>('dashboard')
  const [clients, setClients] = useState<Client[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [orders, setOrders] = useState<Order[]>([])
  const [apiStatus, setApiStatus] = useState<ApiStatus>('checking')
  const [loading, setLoading] = useState(true)
  const [feedback, setFeedback] = useState<{ tone: 'error' | 'success'; message: string } | null>(null)

  const refresh = useCallback(async () => {
    setLoading(true)
    setFeedback(null)
    const health = api.health().then(() => setApiStatus('online')).catch(() => setApiStatus('offline'))
    const [clientResult, productResult, orderResult] = await Promise.allSettled([
      api.clients.list(), api.products.list(), api.orders.list(), health,
    ])
    if (clientResult.status === 'fulfilled') setClients(clientResult.value ?? [])
    if (productResult.status === 'fulfilled') setProducts(productResult.value ?? [])
    if (orderResult.status === 'fulfilled') setOrders(orderResult.value ?? [])
    if ([clientResult, productResult, orderResult].some((result) => result.status === 'rejected')) {
      setFeedback({ tone: 'error', message: 'Não foi possível carregar todos os dados. Confirme se a API está ativa.' })
    }
    setLoading(false)
  }, [])

  useEffect(() => {
    const initialLoad = window.setTimeout(() => void refresh(), 0)
    return () => window.clearTimeout(initialLoad)
  }, [refresh])

  const stats = useMemo(() => ({
    pending: orders.filter((order) => order.status === 'PENDING').length,
    paid: orders.filter((order) => order.status === 'PAID').length,
    canceled: orders.filter((order) => order.status === 'CANCELED').length,
    lowStock: products.filter((product) => product.estoque <= 5).length,
  }), [orders, products])

  async function mutate(action: () => Promise<unknown>, success: string) {
    try {
      setFeedback(null)
      await action()
      setFeedback({ tone: 'success', message: success })
      await refresh()
    } catch (error) {
      setFeedback({ tone: 'error', message: error instanceof Error ? error.message : 'Operação não concluída.' })
    }
  }

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand"><span className="brand-mark">P</span><div><strong>Pedidos</strong><small>control plane</small></div></div>
        <nav aria-label="Navegação principal">
          {navigation.map((item) => (
            <button className={view === item.id ? 'nav-item active' : 'nav-item'} onClick={() => setView(item.id)} key={item.id}>
              <span>{item.marker}</span>{item.label}
            </button>
          ))}
        </nav>
        <div className="service-card">
          <span className={`status-dot ${apiStatus}`} aria-hidden="true" />
          <div><strong>API {apiStatus === 'online' ? 'online' : apiStatus === 'offline' ? 'offline' : 'verificando'}</strong><small>{api.url}</small></div>
        </div>
      </aside>

      <main>
        <header className="topbar">
          <div><span className="eyebrow">OPERAÇÕES / {navigation.find((item) => item.id === view)?.label.toUpperCase()}</span><h1>{navigation.find((item) => item.id === view)?.label}</h1></div>
          <button className="secondary" onClick={() => void refresh()} disabled={loading}>{loading ? 'Atualizando…' : 'Atualizar dados'}</button>
        </header>

        {feedback && <div className={`feedback ${feedback.tone}`} role="status">{feedback.message}</div>}

        {view === 'dashboard' && <Dashboard orders={orders} clients={clients} stats={stats} loading={loading} />}
        {view === 'clientes' && <ClientsView clients={clients} onCreate={(data) => mutate(() => api.clients.create(data), 'Cliente criado com sucesso.')} />}
        {view === 'produtos' && <ProductsView products={products} onCreate={(data) => mutate(() => api.products.create(data), 'Produto criado com sucesso.')} />}
        {view === 'pedidos' && <OrdersView orders={orders} clients={clients} products={products} onCreate={(data) => mutate(() => api.orders.create(data), 'Pedido criado e enviado para processamento.')} onPay={(id) => mutate(() => api.orders.pay(id), 'Pagamento solicitado.')} onCancel={(id) => mutate(() => api.orders.cancel(id), 'Pedido cancelado.')} />}
      </main>
    </div>
  )
}

function Dashboard({ orders, clients, stats, loading }: { orders: Order[]; clients: Client[]; stats: { pending: number; paid: number; canceled: number; lowStock: number }; loading: boolean }) {
  return <div className="stack">
    <section className="metrics-grid" aria-label="Indicadores">
      <Metric label="Pedidos totais" value={orders.length} detail={`${stats.pending} aguardando processamento`} accent />
      <Metric label="Pagamentos concluídos" value={stats.paid} detail={`${stats.canceled} cancelados`} />
      <Metric label="Clientes ativos" value={clients.length} detail="cadastros disponíveis" />
      <Metric label="Estoque crítico" value={stats.lowStock} detail="produtos com até 5 unidades" warning={stats.lowStock > 0} />
    </section>
    <section className="panel split-panel">
      <div><div className="section-heading"><div><span className="eyebrow">FLUXO RECENTE</span><h2>Últimos pedidos</h2></div></div><OrderTable orders={orders.slice(0, 6)} compact loading={loading} /></div>
      <div className="saga-rail"><span className="eyebrow">SAGA DE PAGAMENTO</span><h2>Como o pedido avança</h2>{['Pedido criado e estoque reservado', 'Evento order.created publicado', 'Pagamento processado', 'Confirmação ou compensação'].map((step, index) => <div className="saga-step" key={step}><span>{index + 1}</span><p>{step}</p></div>)}</div>
    </section>
  </div>
}

function Metric({ label, value, detail, accent, warning }: { label: string; value: number; detail: string; accent?: boolean; warning?: boolean }) {
  return <article className={`metric ${accent ? 'accent' : ''} ${warning ? 'warning' : ''}`}><span>{label}</span><strong>{value.toString().padStart(2, '0')}</strong><small>{detail}</small></article>
}

function ClientsView({ clients, onCreate }: { clients: Client[]; onCreate: (data: { name: string; email: string; password: string }) => Promise<void> }) {
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  async function submit(event: FormEvent) { event.preventDefault(); await onCreate(form); setForm({ name: '', email: '', password: '' }) }
  return <div className="content-grid"><section className="panel"><div className="section-heading"><div><span className="eyebrow">BASE CADASTRAL</span><h2>{clients.length} clientes</h2></div></div><div className="data-list">{clients.length ? clients.map((client) => <article className="data-row" key={client.id}><div className="avatar">{client.name.slice(0, 2).toUpperCase()}</div><div><strong>{client.name}</strong><small>{client.email}</small></div><code>{shortId(client.id)}</code></article>) : <Empty message="Nenhum cliente cadastrado." />}</div></section><FormPanel title="Novo cliente" subtitle="Crie uma identidade para novos pedidos"><form onSubmit={(event) => void submit(event)}><Field label="Nome"><input required value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></Field><Field label="E-mail"><input required type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} /></Field><Field label="Senha"><input required type="password" minLength={6} value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} /></Field><button className="primary" type="submit">Cadastrar cliente</button></form></FormPanel></div>
}

function ProductsView({ products, onCreate }: { products: Product[]; onCreate: (data: Product) => Promise<void> }) {
  const [form, setForm] = useState({ id: '', nome: '', preco: '', estoque: '' })
  async function submit(event: FormEvent) { event.preventDefault(); await onCreate({ id: form.id, nome: form.nome, preco: Number(form.preco), estoque: Number(form.estoque) }); setForm({ id: '', nome: '', preco: '', estoque: '' }) }
  return <div className="content-grid"><section className="panel"><div className="section-heading"><div><span className="eyebrow">CATÁLOGO E ESTOQUE</span><h2>{products.length} produtos</h2></div></div><div className="product-grid">{products.length ? products.map((product) => <article className="product-card" key={product.id}><div><code>{product.id}</code><span className={product.estoque <= 5 ? 'stock low' : 'stock'}>{product.estoque} un.</span></div><h3>{product.nome}</h3><strong>{formatMoney(product.preco)}</strong></article>) : <Empty message="Nenhum produto cadastrado." />}</div></section><FormPanel title="Novo produto" subtitle="Cadastre preço e estoque inicial"><form onSubmit={(event) => void submit(event)}><Field label="SKU"><input required value={form.id} onChange={(event) => setForm({ ...form, id: event.target.value })} /></Field><Field label="Nome"><input required value={form.nome} onChange={(event) => setForm({ ...form, nome: event.target.value })} /></Field><div className="field-pair"><Field label="Preço"><input required type="number" min="0.01" step="0.01" value={form.preco} onChange={(event) => setForm({ ...form, preco: event.target.value })} /></Field><Field label="Estoque"><input required type="number" min="0" value={form.estoque} onChange={(event) => setForm({ ...form, estoque: event.target.value })} /></Field></div><button className="primary" type="submit">Cadastrar produto</button></form></FormPanel></div>
}

function OrdersView({ orders, clients, products, onCreate, onPay, onCancel }: { orders: Order[]; clients: Client[]; products: Product[]; onCreate: (data: { cliente_id: string; itens: { produto_id: string; quantidade: number }[] }) => Promise<void>; onPay: (id: string) => Promise<void>; onCancel: (id: string) => Promise<void> }) {
  const [form, setForm] = useState({ clientId: '', productId: '', quantity: '1' })
  async function submit(event: FormEvent) { event.preventDefault(); await onCreate({ cliente_id: form.clientId, itens: [{ produto_id: form.productId, quantidade: Number(form.quantity) }] }); setForm({ ...form, productId: '', quantity: '1' }) }
  return <div className="content-grid orders-layout"><section className="panel"><div className="section-heading"><div><span className="eyebrow">PROCESSAMENTO</span><h2>{orders.length} pedidos</h2></div></div><OrderTable orders={orders} onPay={onPay} onCancel={onCancel} /></section><FormPanel title="Novo pedido" subtitle="Reserve estoque e inicie a Saga"><form onSubmit={(event) => void submit(event)}><Field label="Cliente"><select required value={form.clientId} onChange={(event) => setForm({ ...form, clientId: event.target.value })}><option value="">Selecione</option>{clients.map((client) => <option value={client.id} key={client.id}>{client.name}</option>)}</select></Field><Field label="Produto"><select required value={form.productId} onChange={(event) => setForm({ ...form, productId: event.target.value })}><option value="">Selecione</option>{products.filter((product) => product.estoque > 0).map((product) => <option value={product.id} key={product.id}>{product.nome} · {product.estoque} un.</option>)}</select></Field><Field label="Quantidade"><input required type="number" min="1" value={form.quantity} onChange={(event) => setForm({ ...form, quantity: event.target.value })} /></Field><button className="primary" type="submit" disabled={!clients.length || !products.length}>Criar pedido</button></form></FormPanel></div>
}

function OrderTable({ orders, compact, loading, onPay, onCancel }: { orders: Order[]; compact?: boolean; loading?: boolean; onPay?: (id: string) => Promise<void>; onCancel?: (id: string) => Promise<void> }) {
  if (loading) return <Empty message="Carregando pedidos…" />
  if (!orders.length) return <Empty message="Nenhum pedido encontrado." />
  return <div className="table-wrap"><table><thead><tr><th>Pedido</th><th>Cliente</th><th>Status</th><th>Data</th>{!compact && <th>Ações</th>}</tr></thead><tbody>{orders.map((order) => <tr key={order.id}><td><code title={order.id}>{shortId(order.id)}</code></td><td><code title={order.cliente_id}>{shortId(order.cliente_id)}</code></td><td><span className={`badge ${order.status.toLowerCase()}`}>{order.status}</span></td><td>{formatDate(order.created_at)}</td>{!compact && <td><div className="actions"><button disabled={order.status !== 'PENDING'} onClick={() => void onPay?.(order.id)}>Pagar</button><button disabled={order.status !== 'PENDING'} onClick={() => void onCancel?.(order.id)}>Cancelar</button></div></td>}</tr>)}</tbody></table></div>
}

function FormPanel({ title, subtitle, children }: { title: string; subtitle: string; children: React.ReactNode }) { return <aside className="form-panel"><span className="eyebrow">AÇÃO RÁPIDA</span><h2>{title}</h2><p>{subtitle}</p>{children}</aside> }
function Field({ label, children }: { label: string; children: React.ReactNode }) { return <label className="field"><span>{label}</span>{children}</label> }
function Empty({ message }: { message: string }) { return <div className="empty"><span>—</span><p>{message}</p></div> }

export default App
