package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"pedidos/internal/repository"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
)

// ProductController gerencia o fluxo de dados entre as requisições HTTP e a tabela de produtos no banco
type ProductController struct {
	service *service.ProductService
}

// NewProductController constrói o controlador injetando a dependência do banco de dados
func NewProductController(service *service.ProductService) *ProductController {
	return &ProductController{service: service}
}

// Create lida com a rota POST /produtos para inserir novos itens no estoque
func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Instancia a estrutura que o sqlc gerou para os parâmetros de INSERT
	var input struct {
		ID      string  `json:"id"`
		Nome    string  `json:"nome"`
		Preco   float64 `json:"preco"`
		Estoque int     `json:"estoque"`
	}

	// 2. Transforma o JSON enviado no corpo da requisição do Postman para a nossa estrutura Go
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// Se o JSON estiver mal formatado ou com tipos errados, devolve erro 400 (Requisição Inválida)
		http.Error(w, "Dados inválidos: verifique a estrutura do seu JSON", http.StatusBadRequest)
		return
	}

	// 3. Executa a query de INSERT passando o contexto da requisição e os dados limpos
	produtoSalvo, err := c.service.Create(r.Context(), input.ID, input.Nome, input.Preco, input.Estoque)
	if err != nil {
		// Caso ocorra uma falha de conexão ou restrição no banco, devolve erro 500
		if errors.Is(err, repository.ErrConflict) {
			http.Error(w, "Produto já cadastrado", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// 4. Configura o cabeçalho da resposta avisando que o retorno é um JSON
	w.Header().Set("Content-Type", "application/json")

	// 5. Define o status HTTP como 201 (Created - Criado com sucesso)
	w.WriteHeader(http.StatusCreated)

	// 6. Codifica os dados do produto que o banco acabou de salvar e envia de volta ao Postman
	json.NewEncoder(w).Encode(produtoSalvo)
}

// List lida com a rota GET /produtos para exibir a vitrine do sistema
func (c *ProductController) List(w http.ResponseWriter, r *http.Request) {
	// 1. Executa a busca mapeada de "SELECT * FROM produtos ORDER BY nome ASC"
	produtos, err := c.service.List(r.Context())
	if err != nil {
		// Se a busca falhar, retorna erro interno do servidor
		http.Error(w, "Erro ao buscar a lista de produtos no banco de dados", http.StatusInternalServerError)
		return
	}

	// 2. Configura a resposta como formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 3. Retorna a lista completa de produtos
	json.NewEncoder(w).Encode(produtos)
}

// GetByID lida com a rota GET /produtos/{id} (EXIGIDO NO ENUNCIADO DO DESAFIO)
func (c *ProductController) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Extrai o ID da URL (Como no banco é VARCHAR(50), pegamos diretamente como string)
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		http.Error(w, "ID do produto inválido", http.StatusBadRequest) // Status 400
		return
	}

	// 2. Executa a busca SQL no banco via sqlc
	produto, err := c.service.GetByID(r.Context(), idParam)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "Produto não encontrado", http.StatusNotFound) // Status 404
		return
	}
	if err != nil {
		http.Error(w, "Erro ao buscar produto", http.StatusInternalServerError)
		return
	}

	// 3. Retorna o produto encontrado em formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(produto)
}
