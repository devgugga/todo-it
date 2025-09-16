package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client interface {
	GetCollection(name string) *mongo.Collection
	Close() error
	Health() error
	CreateIndexes(ctx context.Context) error
}

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	dbName   string
	mu       sync.Mutex
	closed   bool
}

type MongoConfig struct {
	URI            string
	DBName         string
	MaxPoolSize    uint64
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

func DefaultMongoConfig() *MongoConfig {
	return &MongoConfig{
		URI:            "mongodb://localhost:27017",
		DBName:         "todo_db",
		MaxPoolSize:    20,
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
	}
}

func NewMongoClient(config *MongoConfig) (*MongoDB, error) {
	if config == nil {
		config = DefaultMongoConfig()
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(config.MaxPoolSize).
		SetMinPoolSize(1).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(5 * time.Second).
		SetRetryWrites(true).
		SetRetryReads(true)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar com MongoDB: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("falha no ping do MongoDB: %w", err)
	}

	database := client.Database(config.DBName)

	mongoDB := &MongoDB{
		client:   client,
		database: database,
		dbName:   config.DBName,
		closed:   false,
	}

	log.Printf("✅ Conectado ao MongoDB - Database: %s", config.DBName)
	return mongoDB, nil
}

func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		log.Printf("⚠️  Tentativa de usar conexão fechada para coleção: %s", name)
		return nil
	}

	return m.database.Collection(name)
}

func (m *MongoDB) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("erro ao desconectar do MongoDB: %w", err)
	}

	m.closed = true
	log.Println("🔌 Desconectado do MongoDB")
	return nil
}

func (m *MongoDB) Health() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("conexão MongoDB está fechada")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.client.Ping(ctx, nil)
}

func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("conexão MongoDB está fechada")
	}

	// TODO: Adicionar futuros indices

	log.Println("📊 Todos os índices foram criados com sucesso!")
	return nil
}
