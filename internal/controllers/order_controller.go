package controllers

import (
	"encoding/json"
	"net/http"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
)

// OrderController segura a dependência do Service
type OrderController struct {
	orderService *service.OrderService
}

// NewOrderController é o construtor do controller de pedidos
func NewOrderController(os *service.OrderService) *OrderController {
	return &OrderController{orderService: os}
}

// CreateOrderRequest representa o JSON esperado no POST /pedidos
type CreateOrderRequest struct {
	IDPedido   string `json:"id_pedido"`
	Cliente    string `json:"cliente"`
	IDProduto  string `json:"id_produto"`
	Quantidade int    `json:"quantidade"`
}

// Create processa a criação de um novo pedido
func (c *OrderController) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "formato de json inválido", http.StatusBadRequest)
		return
	}

	// Aciona o Maestro (Service) que já faz toda a mágica de validar estoque e salvar
	pedido, err := c.orderService.CriarPedido(req.IDPedido, req.Cliente, req.IDProduto, req.Quantidade)
	if err != nil {
		// Se o produto não existir ou o estoque falhar, o erro do seu domínio volta aqui
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pedido)
}

// Pay processa o pagamento de um pedido
func (c *OrderController) Pay(w http.ResponseWriter, r *http.Request) {
	// O chi extrai o {id} direto da URL
	id := chi.URLParam(r, "id")

	if err := c.orderService.PagarPedido(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "Pagamento aprovado com sucesso!"}`))
}

// Cancel processa o cancelamento e estorno de um pedido
func (c *OrderController) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := c.orderService.CancelarPedido(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "Pedido cancelado e estoque devolvido com sucesso!"}`))
}
