package main

import (
	"log"
	"net/http"

	"pedidos/internal/controllers"
	"pedidos/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Instanciamos o "banco de dados" em memória para os produtos
	produtoRepo := repository.NewMemoryProductRepository()

	// 2. Instanciamos o Controller, entregando o repositório para ele trabalhar
	productController := controllers.NewProductController(produtoRepo)

	// 3. Configuração do Roteador (Chi)
	r := chi.NewRouter()

	// Middlewares globais para termos logs bonitos no terminal e proteção contra crash
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 4. Registramos as rotas da nossa API Web
	r.Post("/produtos", productController.Create) // Rota para criar
	r.Get("/produtos", productController.List)    // Rota para listar

	// 5. Liga o servidor HTTP e avisa no terminal
	log.Println("🚀 Sistema de Pedidos (Web) rodando em http://localhost:8080")
	log.Println("Rotas disponíveis:")
	log.Println("  POST /produtos -> Cadastrar novo produto no estoque")
	log.Println("  GET  /produtos -> Listar todos os produtos")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
