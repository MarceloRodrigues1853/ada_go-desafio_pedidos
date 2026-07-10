package controllers

import (
	"encoding/json"
	"net/http"

	"pedidos/internal/repository/db"
)

// ProductController agora depende das queries do banco de dados (sqlc) em vez da memória
type ProductController struct {
	queries *db.Queries
}

// NewProductController injeta as conexões do banco de dados no controlador
func NewProductController(queries *db.Queries) *ProductController {
	return &ProductController{
		queries: queries,
	}
}

// List lida com a rota GET /produtos
func (c *ProductController) List(w http.ResponseWriter, r *http.Request) {
	// 1. Chama a função gerada pelo sqlc que executa "SELECT * FROM produtos ORDER BY nome ASC"
	produtos, err := c.queries.ListProdutos(r.Context())
	if err != nil {
		// Se o banco cair ou a consulta falhar, devolvemos erro 500 (Erro Interno do Servidor)
		http.Error(w, "Erro ao buscar produtos no banco de dados", http.StatusInternalServerError)
		return
	}

	// 2. Prepara a resposta avisando que o formato será JSON
	w.Header().Set("Content-Type", "application/json")

	// 3. Converte a lista de produtos vinda do PostgreSQL para JSON e envia ao Postman
	json.NewEncoder(w).Encode(produtos)
}

// Create lida com a rota POST /produtos e salva diretamente no PostgreSQL
func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Instancia a estrutura que o próprio sqlc gerou (CreateProdutoParams)
	// Como foi configurado 'emit_json_tags: true' no sqlc.yaml, ela já entende o JSON do Postman!
	var input db.CreateProdutoParams

	// 2. Lê o JSON que vem da requisição
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Dados inválidos: verifique o formato do JSON", http.StatusBadRequest)
		return
	}

	// 3. Aciona a query de INSERT no banco de dados
	produtoSalvo, err := c.queries.CreateProduto(r.Context(), input)
	if err != nil {
		http.Error(w, "Erro ao salvar produto no banco de dados", http.StatusInternalServerError)
		return
	}

	// 4. Devolve o Status 201 (Criado) e os dados do produto recém-cadastrado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(produtoSalvo)
}
