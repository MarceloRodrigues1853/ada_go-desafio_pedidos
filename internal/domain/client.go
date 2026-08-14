package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Cliente representa a entidade do banco de dados e do domínio
type Cliente struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// NovoCliente é a fábrica que valida os dados e gera o hash da senha usando bcrypt
func NovoCliente(name, email, password string) (*Cliente, error) {
	// 1. Validação de Invariantes: Nome, e-mail e senha não podem ser vazios
	if name == "" {
		return nil, errors.New("o nome do cliente é obrigatório")
	}
	if email == "" {
		return nil, errors.New("o e-mail do cliente é obrigatório")
	}
	if password == "" {
		return nil, errors.New("a senha do cliente é obrigatória")
	}

	// 2. Gera o Hash da senha utilizando o algoritmo bcrypt (segurança)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// Retorna o erro caso o algoritmo falhe na criptografia
		return nil, err
	}

	// 3. Retorna a instância do Cliente devidamente preenchida e criptografada
	return &Cliente{
		ID:           uuid.New(),       // Gera um novo UUID único para o cliente
		Name:         name,             // Atribui o nome informado
		Email:        email,            // Atribui o e-mail informado
		PasswordHash: string(hash),     // Armazena a senha criptografada (nunca em texto puro)
		CreatedAt:    time.Now().UTC(), // Registra o momento exato da criação em UTC
	}, nil
}
