package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type FornecedoresEmPrd struct {
	ID         string  `json:"id"`
	PrecoVenda float64 `json:"preco_venda"`
}

type ProdutosEmFornecedor struct {
	ID         string  `json:"id"`
	Nome       string  `json:"nome"`
	PrecoVenda float64 `json:"preco_venda"`
	Enabled    bool    `json:"enabled"`
}

type ProdutosEmCategorias struct {
	ID      primitive.ObjectID `bson:"id" json:"id"`
	Nome    string             `bson:"nome" json:"nome"`
	Enabled bool               `bson:"enabled" json:"enabled"`
}

type ProdutosPayload struct {
	Produtos []ProdutosEmFornecedor `json:"produtos"`
}

type FornecedorPaylaod struct {
	Fornecedores []FornecedoresEmPrd `json:"fornecedores"`
}
