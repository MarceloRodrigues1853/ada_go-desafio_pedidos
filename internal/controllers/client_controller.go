package controllers

import (
	"encoding/json"
	"net/http"

	"pedidos/internal/domain"
	"pedidos/internal/repository/db"
)

type ClientController struct {
	queries *db.Queries
}

// NewClientController cria uma nova instância do controller injetando as queries do banco
func NewClientController(queries *db.Queries) *ClientController {
	return &ClientController{
		queries: queries,
	}
}

// Create lida com a rota POST /clientes
func (c *ClientController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Lê o JSON que vem do Postman
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	// 2. Chama o Domínio para aplicar as regras e fazer o Hash da Senha
	novoCliente, err := domain.NovoCliente(input.Name, input.Email, input.Password)
	if err != nil {
		http.Error(w, "Erro ao processar cliente", http.StatusInternalServerError)
		return
	}

	// 3. Salva no PostgreSQL usando as funções geradas pelo sqlc
	clienteSalvo, err := c.queries.CreateCliente(r.Context(), db.CreateClienteParams{
		Name:         novoCliente.Name,
		Email:        novoCliente.Email,
		PasswordHash: novoCliente.PasswordHash,
	})

	if err != nil {
		// Se der erro (ex: email já existe), devolvemos o 409 Conflict exigido no desafio
		http.Error(w, "Email já cadastrado ou erro no banco", http.StatusConflict)
		return
	}

	// 4. Devolve o Status 201 e o cliente cadastrado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(clienteSalvo)
}

// List lida com a rota GET /clientes
func (c *ClientController) List(w http.ResponseWriter, r *http.Request) {
	clientes, err := c.queries.ListClientes(r.Context())
	if err != nil {
		http.Error(w, "Erro ao buscar clientes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientes)
}
