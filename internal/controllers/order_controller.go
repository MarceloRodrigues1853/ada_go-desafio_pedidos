package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pedidos/internal/repository/db"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// OrderController coordena as requisições HTTP para a entidade Pedido
type OrderController struct {
	service *service.OrderService
}

// NewOrderController cria uma nova instância do controller injetando a camada de serviço
func NewOrderController(service *service.OrderService) *OrderController {
	return &OrderController{
		service: service,
	}
}

// Create lida com a rota POST /pedidos
func (c *OrderController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Estrutura temporária para decodificar o payload JSON vindo da requisição
	var input struct {
		ClienteID string `json:"cliente_id"`
		Itens     []struct {
			ProdutoID     string  `json:"produto_id"`
			Quantidade    int32   `json:"quantidade"`
			PrecoUnitario float64 `json:"preco_unitario"`
		} `json:"itens"`
	}

	// Tenta fazer o parse do JSON do corpo da requisição
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Erro de sintaxe no JSON", http.StatusBadRequest) // 400 Dados Inválidos
		return
	}

	// 2. Converte o cliente_id recebido como texto para o tipo UUID
	clienteUUID, err := uuid.Parse(input.ClienteID)
	if err != nil {
		http.Error(w, "O ID do cliente não é um UUID válido", http.StatusBadRequest) // 400 Dados Inválidos
		return
	}

	// 3. Mapeia a lista recebida no JSON para a struct gerada pelo sqlc
	var itensParaSalvar []db.CreateItemPedidoParams
	for _, item := range input.Itens {
		itensParaSalvar = append(itensParaSalvar, db.CreateItemPedidoParams{
			ProdutoID:     item.ProdutoID,
			Quantidade:    item.Quantidade,
			PrecoUnitario: item.PrecoUnitario,
		})
	}

	// 4. Executa o caso de uso de criação no Service
	pedidoSalvo, err := c.service.Create(r.Context(), clienteUUID, itensParaSalvar)
	if err != nil {
		// Trata erro de estoque ou regra de negócio retornando Status 409 Conflict (conforme regra do professor)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// 5. Retorna status 201 Created com o pedido cadastrado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pedidoSalvo)
}

// GetByID lida com a rota GET /pedidos/{id} (NOVO - Exigido no enunciado)
func (c *OrderController) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Extrai o parâmetro {id} da URL
	idParam := chi.URLParam(r, "id")
	pedidoUUID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "ID do pedido inválido", http.StatusBadRequest) // 400
		return
	}

	// 2. Busca o pedido pelo ID no service
	pedido, err := c.service.GetByID(r.Context(), pedidoUUID)
	if err != nil {
		http.Error(w, "Pedido não encontrado", http.StatusNotFound) // 404 Not Found
		return
	}

	// 3. Retorna o pedido encontrado em formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pedido)
}

// ListPaginado lida com a rota GET /pedidos?limit=10&offset=0 (NOVO - Paginação exigida)
func (c *OrderController) ListPaginado(w http.ResponseWriter, r *http.Request) {
	// 1. Lê os parâmetros de query 'limit' e 'offset' da URL (com valores padrão 10 e 0)
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10 // Padrão
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Padrão
	}

	// 2. Busca a lista paginada na camada de serviço
	pedidos, err := c.service.ListPaginado(r.Context(), int32(limit), int32(offset))
	if err != nil {
		http.Error(w, "Erro ao buscar pedidos paginados", http.StatusInternalServerError) // 500
		return
	}

	// 3. Retorna a lista em formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pedidos)
}

// Pay lida com a rota POST /pedidos/{id}/pagar (Ajustado verbo para POST)
func (c *OrderController) Pay(w http.ResponseWriter, r *http.Request) {
	// 1. Extrai o ID da URL
	idParam := chi.URLParam(r, "id")
	pedidoUUID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "ID do pedido inválido", http.StatusBadRequest) // 400
		return
	}

	// 2. Invoca o serviço de pagamento
	if err := c.service.Pay(r.Context(), pedidoUUID); err != nil {
		// Retorna 409 Conflict se o pedido já estiver pago/cancelado (conforme exigido pelo professor)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// 3. Retorna mensagem de sucesso com Status 200 OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "Pedido pago com sucesso"}`))
}

// Cancel lida com a rota POST /pedidos/{id}/cancelar (Ajustado verbo para POST)
func (c *OrderController) Cancel(w http.ResponseWriter, r *http.Request) {
	// 1. Extrai o ID da URL
	idParam := chi.URLParam(r, "id")
	pedidoUUID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "ID do pedido inválido", http.StatusBadRequest) // 400
		return
	}

	// 2. Invoca o serviço de cancelamento
	if err := c.service.Cancel(r.Context(), pedidoUUID); err != nil {
		// Retorna 409 Conflict se o pedido já foi pago ou cancelado anteriormente
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// 3. Retorna confirmação de cancelamento
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "Pedido cancelado com sucesso"}`))
}
