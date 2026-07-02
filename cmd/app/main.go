package main

import (
	"fmt"
	"log"
	"pedidos/internal/domain"
)

func main() {
	produto, err := domain.NovoProduto("1", "Produto 1", 10.0, 10)
	if err != nil {
		log.Fatalf("Erro ao criar produto: %v", err)
	}
	fmt.Println(produto)
}
