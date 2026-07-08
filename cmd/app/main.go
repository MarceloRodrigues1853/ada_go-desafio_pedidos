package main

import (
	"log"
	"net/http"

	"pedidos/internal/controllers"
	"pedidos/internal/repository"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Instanciamos os bancos de dados em memória
	produtoRepo := repository.NewMemoryProductRepository()
	pedidoRepo := repository.NewMemoryOrderRepository()

	// 2. Instanciamos a camada de Serviço (Regras de Negócio orquestradas)
	orderService := service.NewOrderService(produtoRepo, pedidoRepo)

	// 3. Instanciamos os Controllers
	productController := controllers.NewProductController(produtoRepo)
	orderController := controllers.NewOrderController(orderService)

	// 4. Configuração do Roteador (Chi)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 5. Rotas de Produtos
	r.Post("/produtos", productController.Create)
	r.Get("/produtos", productController.List)

	// 6. Rotas de Pedidos
	r.Post("/pedidos", orderController.Create)
	r.Put("/pedidos/{id}/pagar", orderController.Pay)
	r.Put("/pedidos/{id}/cancelar", orderController.Cancel)

	// 7. Liga o servidor HTTP
	log.Println("🚀 Sistema de Pedidos (Web) rodando em http://localhost:8080")
	log.Println("--- ROTAS DISPONÍVEIS ---")
	log.Println("PRODUTOS:")
	log.Println("  POST /produtos")
	log.Println("  GET  /produtos")
	log.Println("PEDIDOS:")
	log.Println("  POST /pedidos")
	log.Println("  PUT  /pedidos/{id}/pagar")
	log.Println("  PUT  /pedidos/{id}/cancelar")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
