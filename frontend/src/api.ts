export type ApiStatus = 'online' | 'offline' | 'checking'

export interface Client {
  id: string
  name: string
  email: string
  created_at: string
}

export interface Product {
  id: string
  nome: string
  preco: number
  estoque: number
}

export type OrderStatus = 'PENDING' | 'PAID' | 'CANCELED'
export type PaymentMethod = 'CARD' | 'PIX' | 'BOLETO'
export type SimulationOutcome = 'APPROVED' | 'DECLINED'

export interface CreateOrderInput {
  cliente_id: string
  payment_method: PaymentMethod
  simulation_outcome: SimulationOutcome
  itens: { produto_id: string; quantidade: number }[]
}

export interface Order {
  id: string
  cliente_id: string
  status: OrderStatus
  created_at: string
}

export interface RAGSource {
  file: string
  position: number
  similarity: number
  content: string
}

export interface RAGResponse {
  answer: string
  has_evidence: boolean
  sources: RAGSource[]
  threshold: number
  documents_indexed: number
  chunks_indexed: number
}

const API_URL = (import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, '')
const RAG_API_URL = (import.meta.env.VITE_RAG_API_URL || 'http://localhost:8081').replace(/\/$/, '')

async function requestFrom<T>(baseUrl: string, path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${baseUrl}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || `A API respondeu com status ${response.status}.`)
  }

  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

const request = <T>(path: string, options?: RequestInit) => requestFrom<T>(API_URL, path, options)

export const api = {
  url: API_URL,
  health: () => request<{ status: string }>('/health'),
  clients: {
    list: () => request<Client[]>('/clientes'),
    create: (data: { name: string; email: string; password: string }) =>
      request<Client>('/clientes', { method: 'POST', body: JSON.stringify(data) }),
  },
  products: {
    list: () => request<Product[]>('/produtos'),
    create: (data: Product) =>
      request<Product>('/produtos', { method: 'POST', body: JSON.stringify(data) }),
  },
  orders: {
    list: () => request<Order[]>('/pedidos?limit=100&offset=0'),
    create: (data: CreateOrderInput) =>
      request<Order>('/pedidos', { method: 'POST', body: JSON.stringify(data) }),
  },
}

export const ragApi = {
  url: RAG_API_URL,
  health: () => requestFrom<{ status: string }>(RAG_API_URL, '/health'),
  ask: (question: string) => requestFrom<RAGResponse>(RAG_API_URL, '/ask', {
    method: 'POST',
    body: JSON.stringify({ question }),
  }),
}
