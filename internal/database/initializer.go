package database

import (
	"context"
	"fmt"
	"log"
	"time"
)

// InitializeDatabase inicializa completamente o banco de dados
func InitializeDatabase(config *MongoConfig) (Client, error) {
	log.Println("🚀 Inicializando banco de dados...")

	// Conecta ao MongoDB
	client, err := NewMongoClient(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar com MongoDB: %w", err)
	}

	// Context com timeout para operações de inicialização
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Garante que as collections existam
	log.Println("📦 Verificando/criando collections...")
	if err := EnsureCollectionsExist(client, ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("falha ao criar collections: %w", err)
	}

	// Cria todos os índices
	log.Println("📊 Criando índices...")
	if err := CreateAllIndexes(client, ctx); err != nil {
		log.Printf("⚠️  Aviso: Erro ao criar índices: %v", err)
	}

	// Verifica se tudo está funcionando
	if err := client.Health(); err != nil {
		client.Close()
		return nil, fmt.Errorf("falha no health check: %w", err)
	}

	log.Println("✅ Banco de dados inicializado com sucesso!")
	return client, nil
}

// EnsureCollectionsExist garante que as collections existam com as configurações corretas
func EnsureCollectionsExist(client Client, ctx context.Context) error {
	mongoClient := client.(*MongoDB)
	return mongoClient.EnsureCollectionsExist(ctx)
}

// DatabaseStats retorna estatísticas do banco
func DatabaseStats(client Client) (*Stats, error) {
	mongoClient := client.(*MongoDB)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections := mongoClient.GetCollections()

	// Conta documentos
	usersCount, err := collections.Users.CountDocuments(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("erro ao contar users: %w", err)
	}

	todosCount, err := collections.Tasks.CountDocuments(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("erro ao contar todos: %w", err)
	}

	return &Stats{
		UsersCount: usersCount,
		TodosCount: todosCount,
		Collections: []string{
			GetCollectionNames().Users,
			GetCollectionNames().Tasks,
		},
	}, nil
}

// Stats representa estatísticas do banco
type Stats struct {
	UsersCount  int64    `json:"users_count"`
	TodosCount  int64    `json:"todos_count"`
	Collections []string `json:"collections"`
}
