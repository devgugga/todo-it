package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionNames define os nomes das collections
type CollectionNames struct {
	Users string
	Tasks string
}

// GetCollectionNames retorna os nomes das collections
func GetCollectionNames() *CollectionNames {
	return &CollectionNames{
		Users: "users",
		Tasks: "tasks",
	}
}

// Collections agrupa todas as collections do banco
type Collections struct {
	Users *mongo.Collection
	Tasks *mongo.Collection
}

// GetCollections retorna todas as collections configuradas
func (m *MongoDB) GetCollections() *Collections {
	names := GetCollectionNames()

	return &Collections{
		Users: m.GetCollection(names.Users),
		Tasks: m.GetCollection(names.Tasks),
	}
}

// GetCollections método para a interface Client
func GetCollections(client Client) *Collections {
	mongoClient := client.(*MongoDB)
	return mongoClient.GetCollections()
}

// CreateAllIndexes cria todos os índices necessários para as entidades
func (m *MongoDB) CreateAllIndexes(ctx context.Context) error {
	collections := m.GetCollections()

	// Cria índices para Users
	if err := m.createUsersIndexes(ctx, collections.Users); err != nil {
		return fmt.Errorf("erro ao criar índices para users: %w", err)
	}

	// Cria índices para Todos
	if err := m.createTodosIndexes(ctx, collections.Tasks); err != nil {
		return fmt.Errorf("erro ao criar índices para todos: %w", err)
	}

	return nil
}

// CreateAllIndexes método para a interface Client
func CreateAllIndexes(client Client, ctx context.Context) error {
	mongoClient := client.(*MongoDB)
	return mongoClient.CreateAllIndexes(ctx)
}

// createUsersIndexes cria índices específicos para a collection de users
func (m *MongoDB) createUsersIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true).SetName("unique_email_idx"),
		},
		{
			Keys:    map[string]interface{}{"is_active": 1},
			Options: options.Index().SetName("is_active_idx"),
		},
		{
			Keys:    map[string]interface{}{"created_at": -1},
			Options: options.Index().SetName("created_at_desc_idx"),
		},
		{
			Keys: map[string]interface{}{
				"email":     1,
				"is_active": 1,
			},
			Options: options.Index().SetName("email_active_compound_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("falha ao criar índices para users: %w", err)
	}

	return nil
}

// createTodosIndexes cria índices específicos para a collection de todos
func (m *MongoDB) createTodosIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"user_id": 1},
			Options: options.Index().SetName("user_id_idx"),
		},
		{
			Keys: map[string]interface{}{
				"user_id": 1,
				"status":  1,
			},
			Options: options.Index().SetName("user_status_compound_idx"),
		},
		{
			Keys: map[string]interface{}{
				"user_id":    1,
				"created_at": -1,
			},
			Options: options.Index().SetName("user_created_desc_idx"),
		},
		{
			Keys: map[string]interface{}{
				"user_id":     1,
				"is_archived": 1,
			},
			Options: options.Index().SetName("user_archived_idx"),
		},
		{
			Keys:    map[string]interface{}{"status": 1},
			Options: options.Index().SetName("status_idx"),
		},
		{
			Keys:    map[string]interface{}{"priority": 1},
			Options: options.Index().SetName("priority_idx"),
		},
		{
			Keys:    map[string]interface{}{"due_date": 1},
			Options: options.Index().SetName("due_date_idx").SetSparse(true),
		},
		{
			Keys:    map[string]interface{}{"tags": 1},
			Options: options.Index().SetName("tags_idx"),
		},
		{
			Keys: map[string]interface{}{
				"user_id":  1,
				"priority": -1,
				"due_date": 1,
			},
			Options: options.Index().SetName("user_priority_due_idx").SetSparse(true),
		},
		// Índice de texto para busca
		{
			Keys: map[string]interface{}{
				"title":       "text",
				"description": "text",
			},
			Options: options.Index().SetName("text_search_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("falha ao criar índices para todos: %w", err)
	}

	return nil
}

// EnsureCollectionsExist garante que as collections existam com as configurações corretas
func (m *MongoDB) EnsureCollectionsExist(ctx context.Context) error {
	names := GetCollectionNames()

	// Lista collections existentes
	collections, err := m.database.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("erro ao listar collections: %w", err)
	}

	existingCollections := make(map[string]bool)
	for _, name := range collections {
		existingCollections[name] = true
	}

	// Cria collections que não existem
	collectionsToCreate := []string{names.Users, names.Tasks}

	for _, collectionName := range collectionsToCreate {
		if !existingCollections[collectionName] {
			// Configura opções da collection se necessário
			opts := options.CreateCollection()

			// Para todos, podemos configurar validação de schema
			if collectionName == names.Tasks {
				opts = opts.SetValidator(m.getTodoValidator())
			}

			err := m.database.CreateCollection(ctx, collectionName, opts)
			if err != nil {
				return fmt.Errorf("erro ao criar collection %s: %w", collectionName, err)
			}

			fmt.Printf("✅ Collection '%s' criada com sucesso\n", collectionName)
		}
	}

	return nil
}

// getTodoValidator retorna o schema validator para a collection todos
func (m *MongoDB) getTodoValidator() map[string]interface{} {
	return map[string]interface{}{
		"$jsonSchema": map[string]interface{}{
			"bsonType": "object",
			"required": []string{"user_id", "title", "status", "priority", "created_at", "updated_at"},
			"properties": map[string]interface{}{
				"user_id": map[string]interface{}{
					"bsonType":    "objectId",
					"description": "ID do usuário proprietário da tarefa",
				},
				"title": map[string]interface{}{
					"bsonType":    "string",
					"minLength":   1,
					"maxLength":   200,
					"description": "Título da tarefa",
				},
				"description": map[string]interface{}{
					"bsonType":    "string",
					"maxLength":   1000,
					"description": "Descrição da tarefa",
				},
				"status": map[string]interface{}{
					"bsonType":    "string",
					"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
					"description": "Status da tarefa",
				},
				"priority": map[string]interface{}{
					"bsonType":    "string",
					"enum":        []string{"low", "medium", "high", "urgent"},
					"description": "Prioridade da tarefa",
				},
				"due_date": map[string]interface{}{
					"bsonType":    "date",
					"description": "Data de vencimento",
				},
				"tags": map[string]interface{}{
					"bsonType": "array",
					"items": map[string]interface{}{
						"bsonType":  "string",
						"minLength": 1,
						"maxLength": 50,
					},
					"maxItems":    10,
					"description": "Tags da tarefa",
				},
				"is_archived": map[string]interface{}{
					"bsonType":    "bool",
					"description": "Se a tarefa está arquivada",
				},
				"created_at": map[string]interface{}{
					"bsonType":    "date",
					"description": "Data de criação",
				},
				"updated_at": map[string]interface{}{
					"bsonType":    "date",
					"description": "Data de atualização",
				},
				"completed_at": map[string]interface{}{
					"bsonType":    "date",
					"description": "Data de conclusão",
				},
			},
		},
	}
}
