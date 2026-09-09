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

const API_URL = (import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, '')

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
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
