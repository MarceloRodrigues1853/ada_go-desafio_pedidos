package controllers

import (
	"encoding/json"
	"net/http"

	"pedidos/internal/domain"
	"pedidos/internal/repository"
)

// ProductController segura a dependência do repositório em memória
type ProductController struct {
	repo repository.ProductRepository // Usamos a interface, respeitando a Clean Architecture!
}

// NewProductController é o nosso construtor
func NewProductController(repo repository.ProductRepository) *ProductController {
	return &ProductController{repo: repo}
}

// CreateProductRequest representa o JSON que o Postman vai enviar
type CreateProductRequest struct {
	ID      string  `json:"id"`
	Nome    string  `json:"nome"`
	Preco   float64 `json:"preco"`
	Estoque int     `json:"estoque"`
}

// Create processa o POST /produtos
func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	// 1. Lê o JSON que chegou do Postman
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "formato de json inválido", http.StatusBadRequest)
		return
	}

	// 2. Chama as regras de negócio (Blindagem do Domínio)
	produto, err := domain.NovoProduto(req.ID, req.Nome, req.Preco, req.Estoque)
	if err != nil {
		// Se o produto for inválido (ex: preço negativo), devolve o erro (ErrProdutoInvalido)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Salva no banco de dados em memória
	if err := c.repo.Salvar(produto); err != nil {
		http.Error(w, "erro ao salvar produto", http.StatusInternalServerError)
		return
	}

	// 4. Retorna sucesso e devolve o produto criado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(produto)
}

// List processa o GET /produtos
func (c *ProductController) List(w http.ResponseWriter, r *http.Request) {
	// 1. Pega todos os produtos do map em memória
	produtos, err := c.repo.Listar()
	if err != nil {
		http.Error(w, "erro ao listar produtos", http.StatusInternalServerError)
		return
	}

	// 2. Devolve a lista como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(produtos)
}
