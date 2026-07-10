package controllers

import (
	"encoding/json"
	"net/http"

	"pedidos/internal/repository/db"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderController struct {
	service *service.OrderService
}

func NewOrderController(service *service.OrderService) *OrderController {
	return &OrderController{
		service: service,
	}
}

// Create lida com a rota POST /pedidos
func (c *OrderController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Define a estrutura exata que esperamos receber do JSON no Postman
	var input struct {
		ClienteID string `json:"cliente_id"`
		Itens     []struct {
			ProdutoID     string  `json:"produto_id"`
			Quantidade    int32   `json:"quantidade"`
			PrecoUnitario float64 `json:"preco_unitario"`
		} `json:"itens"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Erro de sintaxe no JSON", http.StatusBadRequest)
		return
	}

	// 2. Converte o texto (string) do cliente_id num uuid.UUID real para o Postgres
	clienteUUID, err := uuid.Parse(input.ClienteID)
	if err != nil {
		http.Error(w, "O ID do cliente não é um UUID válido", http.StatusBadRequest)
		return
	}

	// 3. Mapeia a lista do Postman para a lista gerada pelo sqlc
	var itensParaSalvar []db.CreateItemPedidoParams
	for _, item := range input.Itens {
		itensParaSalvar = append(itensParaSalvar, db.CreateItemPedidoParams{
			ProdutoID:     item.ProdutoID,
			Quantidade:    item.Quantidade,
			PrecoUnitario: item.PrecoUnitario,
		})
	}

	// 4. Manda tudo para o Service processar as regras e salvar
	pedidoSalvo, err := c.service.Create(r.Context(), clienteUUID, itensParaSalvar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Sucesso!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pedidoSalvo)
}

// Pay lida com a rota PUT /pedidos/{id}/pagar
func (c *OrderController) Pay(w http.ResponseWriter, r *http.Request) {
	// Extrai o {id} da barra de endereço usando o pacote do Chi Router
	idParam := chi.URLParam(r, "id")
	pedidoUUID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "ID do pedido inválido", http.StatusBadRequest)
		return
	}

	if err := c.service.Pay(r.Context(), pedidoUUID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "Pedido pago com sucesso"}`))
}

// Cancel lida com a rota PUT /pedidos/{id}/cancelar
func (c *OrderController) Cancel(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	pedidoUUID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "ID do pedido inválido", http.StatusBadRequest)
		return
	}

	if err := c.service.Cancel(r.Context(), pedidoUUID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "Pedido cancelado com sucesso"}`))
}
