package main

import (
	"fmt"
	"log"

	// Importamos as pastas para poder usar o que construímos
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

	// 3. Criamos um produto inicial no estoque
	notebook, err := domain.NovoProduto("P001", "Notebook Dell", 4500.00, 10)
	if err != nil {
		log.Fatalf("Erro ao criar produto: %v", err)
	}

	// Salvamos o notebook no repositório de produtos para que ele exista no sistema
	produtoRepo.Salvar(notebook)
	fmt.Printf("\n📦 Produto no estoque: %s | Quantidade: %d\n", notebook.Nome, notebook.Estoque)

	// 4. Simulando uma Compra!
	fmt.Printf("\n[ 🛍️ Processando Nova Compra ... ]\n")

	// Vamos tentar comprar alguns notebooks
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

	// =========================================
	// 🔄 DINAMIZANDO: Testando as Regras de Negócio
	// =========================================

	fmt.Println("\n[ 💳 Processando Pagamento ... ]\n")
	// Vamos usar o nosso Service para pagar o pedido que acabamos de criar
	err = orderService.PagarPedido(pedido.ID)
	if err != nil {
		fmt.Printf("❌ Erro ao pagar: %v\n", err)
	} else {
		// Se deu certo, buscamos o pedido atualizado no banco para ver o novo status
		pedidoAtualizado, _ := pedidoRepo.BuscarPorID(pedido.ID)
		fmt.Printf("✅ Pagamento aprovado! Novo status: %s\n", pedidoAtualizado.Status)
	}

	fmt.Println("\n--------------------------------------")
	fmt.Println("\n--- 🛑 Testando Bloqueio de Estoque ---\n")
	// Vamos tentar comprar 15 notebooks (só restaram 8 no estoque)
	quantidadeAbsurda := 15
	fmt.Printf("Tentando comprar %d unidades...\n", quantidadeAbsurda)

	_, errEstoque := orderService.CriarPedido("PED-999-B", "Marcelo", "P001", quantidadeAbsurda)
	if errEstoque != nil {
		// O sistema DEVE cair aqui e imprimir o seu erro customizado!
		fmt.Printf("\n🛡️  Sistema bloqueou a compra corretamente. Motivo: %v\n", errEstoque)
	}
}
