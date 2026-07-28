package controllers

import (
	"encoding/json"
	"net/http"

	"pedidos/internal/domain"
	"pedidos/internal/repository/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ClientController gerencia a comunicação HTTP referente aos clientes
type ClientController struct {
	queries *db.Queries // Repositório gerado pelo sqlc para consultas PostgreSQL
}

// NewClientController cria uma nova instância do controller injetando as queries do banco
func NewClientController(queries *db.Queries) *ClientController {
	return &ClientController{
		queries: queries,
	}
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

	// 2. Chama o Domínio para aplicar regras e criptografar a senha usando bcrypt
	novoCliente, err := domain.NovoCliente(input.Name, input.Email, input.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Status 400
		return
	}

	// 3. Salva no PostgreSQL usando as funções geradas pelo sqlc
	clienteSalvo, err := c.queries.CreateCliente(r.Context(), db.CreateClienteParams{
		Name:         novoCliente.Name,
		Email:        novoCliente.Email,
		PasswordHash: novoCliente.PasswordHash,
	})

	if err != nil {
		// Se der erro (ex: e-mail duplicado), devolve o Status 409 Conflict exigido pelo desafio
		http.Error(w, "Email já cadastrado ou erro no banco", http.StatusConflict)
		return
	}

	// 4. Devolve o Status 201 Created e os dados do cliente cadastrado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(clienteSalvo)
}

// List lida com a rota GET /clientes para trazer todos os registros
func (c *ClientController) List(w http.ResponseWriter, r *http.Request) {
	clientes, err := c.queries.ListClientes(r.Context())
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
	cliente, err := c.queries.GetCliente(r.Context(), clienteUUID)
	if err != nil {
		http.Error(w, "Cliente não encontrado", http.StatusNotFound) // Status 404
		return
	}

	// 3. Retorna o cliente em formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cliente)
}
