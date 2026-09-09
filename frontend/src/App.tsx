import { type FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { api, ragApi, type ApiStatus, type Client, type CreateOrderInput, type Order, type PaymentMethod, type Product, type RAGResponse, type SimulationOutcome } from './api'
import './App.css'

type View = 'dashboard' | 'clientes' | 'produtos' | 'pedidos' | 'tutor'

const navigation: { id: View; label: string; marker: string }[] = [
  { id: 'dashboard', label: 'Visão geral', marker: '01' },
  { id: 'clientes', label: 'Clientes', marker: '02' },
  { id: 'produtos', label: 'Produtos', marker: '03' },
  { id: 'pedidos', label: 'Pedidos', marker: '04' },
  { id: 'tutor', label: 'Tutor RAG', marker: '05' },
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

  const refresh = useCallback(async (silent = false) => {
    if (!silent) {
      setLoading(true)
      setFeedback(null)
    }
    const health = api.health().then(() => setApiStatus('online')).catch(() => setApiStatus('offline'))
    const [clientResult, productResult, orderResult] = await Promise.allSettled([
      api.clients.list(), api.products.list(), api.orders.list(), health,
    ])
    if (clientResult.status === 'fulfilled') setClients(clientResult.value ?? [])
    if (productResult.status === 'fulfilled') setProducts(productResult.value ?? [])
    if (orderResult.status === 'fulfilled') setOrders(orderResult.value ?? [])
    if (!silent && [clientResult, productResult, orderResult].some((result) => result.status === 'rejected')) {
      setFeedback({ tone: 'error', message: 'Não foi possível carregar todos os dados. Confirme se a API está ativa.' })
    }
    if (!silent) setLoading(false)
  }, [])

  useEffect(() => {
    const initialLoad = window.setTimeout(() => void refresh(), 0)
    return () => window.clearTimeout(initialLoad)
  }, [refresh])

  const hasPendingOrders = orders.some((order) => order.status === 'PENDING')

  useEffect(() => {
    if (!hasPendingOrders || apiStatus !== 'online') return
    const polling = window.setInterval(() => void refresh(true), 1500)
    return () => window.clearInterval(polling)
  }, [apiStatus, hasPendingOrders, refresh])

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
          {view !== 'tutor' && <button className="secondary" onClick={() => void refresh()} disabled={loading}>{loading ? 'Atualizando…' : 'Atualizar dados'}</button>}
        </header>

        {feedback && <div className={`feedback ${feedback.tone}`} role="status">{feedback.message}</div>}

        {view === 'dashboard' && <Dashboard orders={orders} clients={clients} stats={stats} loading={loading} />}
        {view === 'clientes' && <ClientsView clients={clients} onCreate={(data) => mutate(() => api.clients.create(data), 'Cliente criado com sucesso.')} />}
        {view === 'produtos' && <ProductsView products={products} onCreate={(data) => mutate(() => api.products.create(data), 'Produto criado com sucesso.')} />}
        {view === 'pedidos' && <OrdersView orders={orders} clients={clients} products={products} onCreate={(data) => mutate(() => api.orders.create(data), data.simulation_outcome === 'APPROVED' ? 'Pedido criado. A aprovação será processada automaticamente pela Saga.' : 'Pedido criado. A recusa será processada automaticamente e o estoque devolvido.')} />}
        {view === 'tutor' && <TutorView />}
      </main>
    </div>
  )
}

function TutorView() {
  const [question, setQuestion] = useState('')
  const [result, setResult] = useState<RAGResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault()
    const normalizedQuestion = question.trim()
    if (!normalizedQuestion) {
      setResult(null)
      setError('Digite uma pergunta antes de consultar a documentação.')
      return
    }
    try {
      setLoading(true)
      setError('')
      setResult(await ragApi.ask(normalizedQuestion))
    } catch (requestError) {
      setResult(null)
      setError(requestError instanceof Error ? requestError.message : 'Não foi possível consultar o Tutor RAG.')
    } finally {
      setLoading(false)
    }
  }

  const examples = [
    'Como funciona o fluxo de pagamento de um pedido?',
    'Qual é a política de férias dos funcionários da empresa?',
  ]

  return <div className="tutor-layout">
    <section className="panel tutor-query">
      <span className="eyebrow">RECUPERAÇÃO DOCUMENTAL</span>
      <h2>Consulte a documentação do projeto</h2>
      <p>O tutor busca evidências em arquivos Markdown autorizados. As fontes ajudam na investigação, mas a implementação deve ser confirmada no código.</p>
      <form onSubmit={(event) => void submit(event)}>
        <label className="field"><span>Pergunta</span><textarea maxLength={500} rows={5} value={question} onChange={(event) => setQuestion(event.target.value)} placeholder="Ex.: Como funciona o fluxo de pagamento?" /></label>
        <div className="query-footer"><small>{question.length}/500 caracteres</small><button className="primary tutor-submit" type="submit" disabled={loading}>{loading ? 'Consultando…' : 'Buscar evidências'}</button></div>
      </form>
      <div className="example-queries"><span>Experimente:</span>{examples.map((example) => <button type="button" onClick={() => { setQuestion(example); setError('') }} key={example}>{example}</button>)}</div>
      <small className="endpoint-note">Serviço: {ragApi.url}</small>
    </section>

    <section className="panel tutor-results" aria-live="polite">
      <div className="section-heading"><div><span className="eyebrow">RESULTADO</span><h2>Evidências recuperadas</h2></div></div>
      {error && <div className="feedback error">{error}</div>}
      {!error && !result && <Empty message="Faça uma pergunta para consultar as fontes indexadas." />}
      {result && <>
        <div className="rag-stats"><span>{result.documents_indexed} documentos</span><span>{result.chunks_indexed} trechos</span><span>limite {result.threshold.toFixed(2)}</span></div>
        {!result.has_evidence ? <div className="no-evidence"><strong>Nenhuma evidência suficiente</strong><p>{result.answer}</p></div> : <div className="source-list">{result.sources.map((source, index) => <article className="source-card" key={`${source.file}-${source.position}-${index}`}><div className="source-meta"><strong>{source.file}</strong><span>Trecho {source.position}</span><span>Similaridade {source.similarity.toFixed(3)}</span></div><p>{source.content}</p></article>)}</div>}
      </>}
    </section>
  </div>
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
      <div><div className="section-heading"><div><span className="eyebrow">FLUXO RECENTE</span><h2>Últimos pedidos</h2></div></div><OrderTable orders={orders.slice(0, 6)} loading={loading} /></div>
      <div className="saga-rail"><span className="eyebrow">SAGA DE PAGAMENTO</span><h2>Como o pedido avança</h2>{['Pedido criado e estoque reservado', 'Evento order.created publicado', 'Pagamento processado', 'Confirmação ou compensação'].map((step, index) => <div className="saga-step" key={step}><span>{index + 1}</span><p>{step}</p></div>)}</div>
    </section>
  </div>
}

function Metric({ label, value, detail, accent, warning }: { label: string; value: number; detail: string; accent?: boolean; warning?: boolean }) {
  return <article className={`metric ${accent ? 'accent' : ''} ${warning ? 'warning' : ''}`}><span>{label}</span><strong>{value.toString().padStart(2, '0')}</strong><small>{detail}</small></article>
}

function ClientsView({ clients, onCreate }: { clients: Client[]; onCreate: (data: { name: string; email: string; password: string }) => Promise<void> }) {
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  const [query, setQuery] = useState('')
  const normalizedQuery = query.trim().toLocaleLowerCase('pt-BR')
  const filteredClients = clients.filter((client) => `${client.name} ${client.email} ${client.id}`.toLocaleLowerCase('pt-BR').includes(normalizedQuery))
  async function submit(event: FormEvent) { event.preventDefault(); await onCreate(form); setForm({ name: '', email: '', password: '' }) }
  return <div className="content-grid"><section className="panel"><div className="section-heading"><div><span className="eyebrow">BASE CADASTRAL</span><h2>{clients.length} clientes</h2></div></div><SearchField label="Buscar clientes" value={query} onChange={setQuery} placeholder="Nome, e-mail ou ID" resultCount={filteredClients.length} totalCount={clients.length} /><div className="data-list">{filteredClients.length ? filteredClients.map((client) => <article className="data-row" key={client.id}><div className="avatar">{client.name.slice(0, 2).toUpperCase()}</div><div><strong>{client.name}</strong><small>{client.email}</small></div><code>{shortId(client.id)}</code></article>) : <Empty message={clients.length ? "Nenhum cliente corresponde à busca." : "Nenhum cliente cadastrado."} />}</div></section><FormPanel title="Novo cliente" subtitle="Crie uma identidade para novos pedidos"><form onSubmit={(event) => void submit(event)}><Field label="Nome"><input required value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></Field><Field label="E-mail"><input required type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} /></Field><Field label="Senha"><input required type="password" minLength={6} value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} /></Field><button className="primary" type="submit">Cadastrar cliente</button></form></FormPanel></div>
}

function ProductsView({ products, onCreate }: { products: Product[]; onCreate: (data: Product) => Promise<void> }) {
  const [form, setForm] = useState({ id: '', nome: '', preco: '', estoque: '' })
  const [query, setQuery] = useState('')
  const normalizedQuery = query.trim().toLocaleLowerCase('pt-BR')
  const filteredProducts = products.filter((product) => `${product.nome} ${product.id}`.toLocaleLowerCase('pt-BR').includes(normalizedQuery))
  async function submit(event: FormEvent) { event.preventDefault(); await onCreate({ id: form.id, nome: form.nome, preco: Number(form.preco), estoque: Number(form.estoque) }); setForm({ id: '', nome: '', preco: '', estoque: '' }) }
  return <div className="content-grid"><section className="panel"><div className="section-heading"><div><span className="eyebrow">CATÁLOGO E ESTOQUE</span><h2>{products.length} produtos</h2></div></div><SearchField label="Buscar produtos" value={query} onChange={setQuery} placeholder="Nome ou SKU" resultCount={filteredProducts.length} totalCount={products.length} /><div className="product-grid">{filteredProducts.length ? filteredProducts.map((product) => <article className="product-card" key={product.id}><div><code>{product.id}</code><span className={product.estoque <= 5 ? 'stock low' : 'stock'}>{product.estoque} un.</span></div><h3>{product.nome}</h3><strong>{formatMoney(product.preco)}</strong></article>) : <Empty message={products.length ? "Nenhum produto corresponde à busca." : "Nenhum produto cadastrado."} />}</div></section><FormPanel title="Novo produto" subtitle="Cadastre preço e estoque inicial"><form onSubmit={(event) => void submit(event)}><Field label="SKU"><input required value={form.id} onChange={(event) => setForm({ ...form, id: event.target.value })} /></Field><Field label="Nome"><input required value={form.nome} onChange={(event) => setForm({ ...form, nome: event.target.value })} /></Field><div className="field-pair"><Field label="Preço"><input required type="number" min="0.01" step="0.01" value={form.preco} onChange={(event) => setForm({ ...form, preco: event.target.value })} /></Field><Field label="Estoque"><input required type="number" min="0" value={form.estoque} onChange={(event) => setForm({ ...form, estoque: event.target.value })} /></Field></div><button className="primary" type="submit">Cadastrar produto</button></form></FormPanel></div>
}

function OrdersView({ orders, clients, products, onCreate }: { orders: Order[]; clients: Client[]; products: Product[]; onCreate: (data: CreateOrderInput) => Promise<void> }) {
  const [form, setForm] = useState<{ clientId: string; productId: string; quantity: string; paymentMethod: PaymentMethod; simulationOutcome: SimulationOutcome }>({ clientId: '', productId: '', quantity: '1', paymentMethod: 'CARD', simulationOutcome: 'APPROVED' })
  const [page, setPage] = useState(1)
  const pageSize = 10
  const pageCount = Math.max(1, Math.ceil(orders.length / pageSize))
  const currentPage = Math.min(page, pageCount)
  const visibleOrders = orders.slice((currentPage - 1) * pageSize, currentPage * pageSize)
  async function submit(event: FormEvent) {
    event.preventDefault()
    await onCreate({
      cliente_id: form.clientId,
      payment_method: form.paymentMethod,
      simulation_outcome: form.simulationOutcome,
      itens: [{ produto_id: form.productId, quantidade: Number(form.quantity) }],
    })
    setForm({ ...form, productId: '', quantity: '1' })
  }
  const hasPendingOrders = orders.some((order) => order.status === 'PENDING')
  return <div className="content-grid orders-layout"><section className="panel"><div className="section-heading"><div><span className="eyebrow">PROCESSAMENTO</span><h2>{orders.length} pedidos</h2></div><span className="live-status">{hasPendingOrders ? 'Saga processando automaticamente' : 'Status controlado pela Saga'}</span></div><OrderTable orders={visibleOrders} /><Pagination page={currentPage} pageCount={pageCount} onPageChange={setPage} /></section><FormPanel title="Novo pedido" subtitle="Reserve estoque e simule o pagamento automático"><form onSubmit={(event) => void submit(event)}><Field label="Cliente"><select required value={form.clientId} onChange={(event) => setForm({ ...form, clientId: event.target.value })}><option value="">Selecione</option>{clients.map((client) => <option value={client.id} key={client.id}>{client.name} · {shortId(client.id)}</option>)}</select></Field><Field label="Produto"><select required value={form.productId} onChange={(event) => setForm({ ...form, productId: event.target.value })}><option value="">Selecione</option>{products.filter((product) => product.estoque > 0).map((product) => <option value={product.id} key={product.id}>{product.nome} · {product.id} · {product.estoque} un.</option>)}</select></Field><Field label="Quantidade"><input required type="number" min="1" value={form.quantity} onChange={(event) => setForm({ ...form, quantity: event.target.value })} /></Field><Field label="Meio de pagamento"><select value={form.paymentMethod} onChange={(event) => setForm({ ...form, paymentMethod: event.target.value as PaymentMethod })}><option value="CARD">Cartão</option><option value="PIX">Pix</option><option value="BOLETO">Boleto</option></select></Field><Field label="Resultado da simulação"><select value={form.simulationOutcome} onChange={(event) => setForm({ ...form, simulationOutcome: event.target.value as SimulationOutcome })}><option value="APPROVED">Aprovar pagamento</option><option value="DECLINED">Recusar pagamento</option></select></Field><div className={`simulation-note ${form.simulationOutcome === 'DECLINED' ? 'declined' : 'approved'}`} role="status"><strong>Pagamento fictício automático</strong><span>{form.simulationOutcome === 'APPROVED' ? 'Ao criar, a Saga processará a aprovação e concluirá como PAID.' : 'Ao criar, a Saga processará a recusa, cancelará o pedido e devolverá o estoque.'}</span></div><button className="primary" type="submit" disabled={!clients.length || !products.length}>{form.simulationOutcome === 'APPROVED' ? 'Criar e simular aprovação' : 'Criar e simular recusa'}</button></form></FormPanel></div>
}

function OrderTable({ orders, loading }: { orders: Order[]; loading?: boolean }) {
  if (loading) return <Empty message="Carregando pedidos…" />
  if (!orders.length) return <Empty message="Nenhum pedido encontrado." />
  return <div className="table-wrap"><table><thead><tr><th>Pedido</th><th>Cliente</th><th>Status</th><th>Data</th></tr></thead><tbody>{orders.map((order) => <tr key={order.id}><td><code title={order.id}>{shortId(order.id)}</code></td><td><code title={order.cliente_id}>{shortId(order.cliente_id)}</code></td><td><span className={`badge ${order.status.toLowerCase()}`}>{order.status}</span></td><td>{formatDate(order.created_at)}</td></tr>)}</tbody></table></div>
}

function FormPanel({ title, subtitle, children }: { title: string; subtitle: string; children: React.ReactNode }) { return <aside className="form-panel"><span className="eyebrow">AÇÃO RÁPIDA</span><h2>{title}</h2><p>{subtitle}</p>{children}</aside> }
function Field({ label, children }: { label: string; children: React.ReactNode }) { return <label className="field"><span>{label}</span>{children}</label> }
function Empty({ message }: { message: string }) { return <div className="empty"><span>—</span><p>{message}</p></div> }

function SearchField({ label, value, onChange, placeholder, resultCount, totalCount }: { label: string; value: string; onChange: (value: string) => void; placeholder: string; resultCount: number; totalCount: number }) {
  return <div className="search-field"><label><span>{label}</span><input type="search" value={value} onChange={(event) => onChange(event.target.value)} placeholder={placeholder} /></label><small>{resultCount} de {totalCount}</small></div>
}

function Pagination({ page, pageCount, onPageChange }: { page: number; pageCount: number; onPageChange: (page: number) => void }) {
  if (pageCount <= 1) return null
  return <nav className="pagination" aria-label="Paginação dos pedidos"><button type="button" disabled={page === 1} onClick={() => onPageChange(page - 1)}>Anterior</button><span>Página {page} de {pageCount}</span><button type="button" disabled={page === pageCount} onClick={() => onPageChange(page + 1)}>Próxima</button></nav>
}

export default App
