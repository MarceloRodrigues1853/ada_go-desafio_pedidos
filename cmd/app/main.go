package main

import (
	"fmt"
	"log"

	// Importamos as nossas pastas para poder usar o que construímos
	"pedidos/internal/domain"
	"pedidos/internal/repository"
	"pedidos/internal/service"
)

func main() {
	fmt.Println("🛒 Iniciando o Sistema de Pedidos...")
	fmt.Println("\n--------------------------------------")

	// 1. Instanciamos os "bancos de dados" em memória
	produtoRepo := repository.NewMemoryProductRepository()
	pedidoRepo := repository.NewMemoryOrderRepository()

	// 2. Instanciamos o Maestro (Service), entregando os repositórios para ele trabalhar
	orderService := service.NewOrderService(produtoRepo, pedidoRepo)

	// 3. Criamos um produto inicial no estoque (usando a Factory do seu professor!)
	notebook, err := domain.NovoProduto("P001", "Notebook Dell", 4500.00, 10)
	if err != nil {
		log.Fatalf("Erro ao criar produto: %v", err)
	}

	// Salvamos o notebook no repositório de produtos para que ele exista no sistema
	produtoRepo.Salvar(notebook)
	fmt.Printf("\n📦 Produto no estoque: %s | Quantidade: %d\n", notebook.Nome, notebook.Estoque)

	// 4. Simulando uma Compra!
	fmt.Printf("\n[    🛍️ Processando Nova Compra ...   ]\n")

	// Vamos tentar comprar 2 notebooks
	quantidadeComprada := 2
	pedido, err := orderService.CriarPedido("PED-999", "Marcelo Rodrigues", "P001", quantidadeComprada)

	if err != nil {
		// Se faltar estoque ou produto não existir, o sistema avisa e para aqui
		log.Fatalf("❌ Falha na compra: %v", err)
	}

	// 5. Exibindo o Sucesso
	fmt.Printf("\n✅ Pedido %s criado com sucesso!\n", pedido.ID)
	fmt.Printf("\n👤 Cliente: %s\n", pedido.Cliente)
	fmt.Printf("📊 Status Atual: %s\n", pedido.Status)

	// Olha o nosso CalcularTotal() sendo chamado na prática!
	fmt.Printf("💰 Total a pagar: R$ %.2f\n", pedido.CalcularTotal())

	// 6. Verificando o estoque após a compra
	produtoAtualizado, _ := produtoRepo.BuscarPorID("P001")
	fmt.Printf("\n📦 Estoque final do %s: %d unidades restantes.\n", produtoAtualizado.Nome, produtoAtualizado.Estoque)
}
