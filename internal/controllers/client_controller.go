package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"pedidos/internal/repository"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ClientController gerencia a comunicação HTTP referente aos clientes
type ClientController struct {
	service *service.ClientService
}

// NewClientController cria uma nova instância do controller injetando as queries do banco
func NewClientController(service *service.ClientService) *ClientController {
	return &ClientController{service: service}
}

// Create lida com a rota POST /clientes para cadastrar novos clientes
func (c *ClientController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Estrutura temporária para decodificar o JSON vindo do cliente HTTP
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Tenta fazer o parse do payload JSON do corpo da requisição
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest) // Status 400
		return
	}

	clienteSalvo, err := c.service.Create(r.Context(), input.Name, input.Email, input.Password)

	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			http.Error(w, "Email já cadastrado", http.StatusConflict)
		} else if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// 4. Devolve o Status 201 Created e os dados do cliente cadastrado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(clienteSalvo)
}

// List lida com a rota GET /clientes para trazer todos os registros
func (c *ClientController) List(w http.ResponseWriter, r *http.Request) {
	clientes, err := c.service.List(r.Context())
	if err != nil {
		http.Error(w, "Erro ao buscar clientes", http.StatusInternalServerError) // Status 500
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(clientes)
}

// GetByID lida com a rota GET /clientes/{id} (EXIGIDO NO ENUNCIADO DO DESAFIO)
func (c *ClientController) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Extrai o parâmetro {id} da URL usando a biblioteca do Chi Router
	idParam := chi.URLParam(r, "id")
	clienteUUID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "ID do cliente não é um UUID válido", http.StatusBadRequest) // Status 400
		return
	}

	// 2. Executa a consulta no PostgreSQL via sqlc
	cliente, err := c.service.GetByID(r.Context(), clienteUUID)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "Cliente não encontrado", http.StatusNotFound) // Status 404
		return
	}
	if err != nil {
		http.Error(w, "Erro ao buscar cliente", http.StatusInternalServerError)
		return
	}

	// 3. Retorna o cliente em formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cliente)
}
