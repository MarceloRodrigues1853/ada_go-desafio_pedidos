package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Cliente representa a tabela clientes no banco de dados
type Cliente struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"` // Regra do desafio: no JSON deve ser passwordHash
	CreatedAt    time.Time `json:"created_at"`
}

// NovoCliente é a fábrica que já cria o cliente fazendo o hash da senha
func NovoCliente(name, email, password string) (*Cliente, error) {
	// 1. Gera o Hash da senha (Regra: Não salvar texto puro)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Cliente{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}, nil
}
