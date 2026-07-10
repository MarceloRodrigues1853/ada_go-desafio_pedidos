package controllers

import (
	"encoding/json"
	"net/http"

	"pedidos/internal/repository/db"
)

// ProductController gerencia o fluxo de dados entre as requisições HTTP e a tabela de produtos no banco
type ProductController struct {
	queries *db.Queries // Ponteiro para as funções geradas pelo sqlc
}

// NewProductController constrói o controlador injetando a dependência do banco de dados
func NewProductController(queries *db.Queries) *ProductController {
	return &ProductController{
		queries: queries, // Conecta as rotas do arquivo main às funções SQL
	}
}

// Create lida com a rota POST /produtos para inserir novos itens no estoque
func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Instancia a estrutura que o sqlc gerou para os parâmetros de INSERT
	var input db.CreateProdutoParams

	// 2. Transforma o JSON enviado no corpo da requisição do Postman para a nossa estrutura Go
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// Se o JSON estiver mal formatado ou com tipos errados, devolve erro 400 (Requisição Inválida)
		http.Error(w, "Dados inválidos: verifique a estrutura do seu JSON", http.StatusBadRequest)
		return
	}

	// 3. Executa a query de INSERT passando o contexto da requisição e os dados limpos
	produtoSalvo, err := c.queries.CreateProduto(r.Context(), input)
	if err != nil {
		// Caso ocorra uma falha de conexão ou restrição no banco, devolve erro 500
		http.Error(w, "Falha ao gravar o produto no PostgreSQL", http.StatusInternalServerError)
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
	produtos, err := c.queries.ListProdutos(r.Context())
	if err != nil {
		// Se a busca falhar, retorna erro interno do servidor
		http.Error(w, "Erro ao buscar a lista de produtos no banco de dados", http.StatusInternalServerError)
		return
	}

	// 2. Configura a resposta como formato JSON
	w.Header().Set("Content-Type", "application/json")

	// 3. Retorna a lista completa com os dados do seed ou novos cadastros
	json.NewEncoder(w).Encode(produtos)
}
