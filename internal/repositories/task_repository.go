package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/devgugga/todo-it/internal/database"
	"github.com/devgugga/todo-it/internal/entities"
	"github.com/devgugga/todo-it/internal/enums"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TodoFilters representa filtros para busca de todos
type TaskFilters struct {
	Status     enums.TaskStatus   `json:"status"`
	Priority   enums.TaskPriority `json:"priority"`
	Tags       []string           `json:"tags"`
	IsArchived *bool              `json:"is_archived"`
	DueBefore  *time.Time         `json:"due_before"`
	DueAfter   *time.Time         `json:"due_after"`
	Search     string             `json:"search"`
}

// TodoStats representa estatísticas dos todos
type TaskStats struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	InProgress int64 `json:"in_progress"`
	Completed  int64 `json:"completed"`
	Cancelled  int64 `json:"cancelled"`
	Archived   int64 `json:"archived"`
	Overdue    int64 `json:"overdue"`
}

// TodoRepository interface define os métodos do repositório de todos
type TodoRepository interface {
	Create(ctx context.Context, todo *entities.Task) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Task, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int64, filters *TaskFilters) ([]*entities.Task, int64, error)
	Update(ctx context.Context, todo *entities.Task) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status enums.TaskStatus) error
	BulkUpdateStatus(ctx context.Context, ids []primitive.ObjectID, status enums.TaskStatus) (int64, error)
	BulkDelete(ctx context.Context, ids []primitive.ObjectID) (int64, error)
	GetStatsByUser(ctx context.Context, userID primitive.ObjectID) (*TaskStats, error)
	GetOverdueTodos(ctx context.Context, userID primitive.ObjectID) ([]*entities.Task, error)
}

// todoRepository implementa TodoRepository
type todoRepository struct {
	collection *mongo.Collection
}

// NewTodoRepository cria uma nova instância do repositório
func NewTodoRepository(db database.Client) TodoRepository {
	collections := database.GetCollections(db)

	return &todoRepository{
		collection: collections.Tasks,
	}
}

// Create cria um novo todo
func (r *todoRepository) Create(ctx context.Context, todo *entities.Task) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	result, err := r.collection.InsertOne(ctx, todo)
	if err != nil {
		return fmt.Errorf("erro ao criar todo: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		todo.ID = oid
	}

	return nil
}

// GetByID busca todo por ID
func (r *todoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Task, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	var todo entities.Task
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("todo não encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar todo: %w", err)
	}

	return &todo, nil
}

// GetByUserID busca todos por usuário com filtros e paginação
func (r *todoRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int64, filters *TaskFilters) ([]*entities.Task, int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	// Constrói filtro base
	filter := bson.M{"user_id": userID}

	// Aplica filtros
	if filters != nil {
		r.applyFilters(filter, filters)
	}

	// Conta total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao contar todos: %w", err)
	}

	// Calcula skip
	skip := (page - 1) * limit

	// Opções de busca
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	// Executa busca
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar todos: %w", err)
	}
	defer cursor.Close(ctx)

	var todos []*entities.Task
	for cursor.Next(ctx) {
		var todo entities.Task
		if err := cursor.Decode(&todo); err != nil {
			return nil, 0, fmt.Errorf("erro ao decodificar todo: %w", err)
		}
		todos = append(todos, &todo)
	}

	return todos, total, nil
}

// applyFilters aplica filtros na query
func (r *todoRepository) applyFilters(filter bson.M, filters *TaskFilters) {
	if filters.Status != "" {
		filter["status"] = filters.Status
	}

	if filters.Priority != "" {
		filter["priority"] = filters.Priority
	}

	if len(filters.Tags) > 0 {
		filter["tags"] = bson.M{"$in": filters.Tags}
	}

	if filters.IsArchived != nil {
		filter["is_archived"] = *filters.IsArchived
	}

	if filters.DueBefore != nil {
		if filter["due_date"] == nil {
			filter["due_date"] = bson.M{}
		}
		filter["due_date"].(bson.M)["$lte"] = *filters.DueBefore
	}

	if filters.DueAfter != nil {
		if filter["due_date"] == nil {
			filter["due_date"] = bson.M{}
		}
		filter["due_date"].(bson.M)["$gte"] = *filters.DueAfter
	}

	if filters.Search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"description": bson.M{"$regex": filters.Search, "$options": "i"}},
		}
	}
}

// Update atualiza um todo
func (r *todoRepository) Update(ctx context.Context, todo *entities.Task) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	todo.PrepareForUpdate()

	filter := bson.M{"_id": todo.ID}
	update := bson.M{
		"$set": bson.M{
			"title":        todo.Title,
			"description":  todo.Description,
			"status":       todo.Status,
			"priority":     todo.Priority,
			"due_date":     todo.DueDate,
			"tags":         todo.Tags,
			"is_archived":  todo.IsArchived,
			"updated_at":   todo.UpdatedAt,
			"completed_at": todo.CompletedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar todo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("todo não encontrado")
	}

	return nil
}

// Delete remove um todo
func (r *todoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("erro ao deletar todo: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("todo não encontrado")
	}

	return nil
}

// UpdateStatus atualiza apenas o status de um todo
func (r *todoRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status enums.TaskStatus) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Se está marcando como concluído, adiciona completed_at
	if status == enums.StatusCompleted {
		update["$set"].(bson.M)["completed_at"] = time.Now()
	} else {
		update["$unset"] = bson.M{"completed_at": ""}
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("todo não encontrado")
	}

	return nil
}

// BulkUpdateStatus atualiza status de múltiplos todos
func (r *todoRepository) BulkUpdateStatus(ctx context.Context, ids []primitive.ObjectID, status enums.TaskStatus) (int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == enums.StatusCompleted {
		update["$set"].(bson.M)["completed_at"] = time.Now()
	} else {
		update["$unset"] = bson.M{"completed_at": ""}
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("erro ao atualizar status em lote: %w", err)
	}

	return result.ModifiedCount, nil
}

// BulkDelete remove múltiplos todos
func (r *todoRepository) BulkDelete(ctx context.Context, ids []primitive.ObjectID) (int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("erro ao deletar em lote: %w", err)
	}

	return result.DeletedCount, nil
}

// GetStatsByUser retorna estatísticas dos todos por usuário
func (r *todoRepository) GetStatsByUser(ctx context.Context, userID primitive.ObjectID) (*TaskStats, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{"user_id": userID},
		},
		{
			"$group": bson.M{
				"_id":   "$status",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter estatísticas: %w", err)
	}
	defer cursor.Close(ctx)

	stats := &TaskStats{}
	statusCounts := make(map[string]int64)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("erro ao decodificar estatística: %w", err)
		}
		statusCounts[result.ID] = result.Count
		stats.Total += result.Count
	}

	// Preenche estatísticas por status
	stats.Pending = statusCounts[string(enums.StatusPending)]
	stats.InProgress = statusCounts[string(enums.StatusInProgress)]
	stats.Completed = statusCounts[string(enums.StatusCompleted)]
	stats.Cancelled = statusCounts[string(enums.StatusCancelled)]

	// Conta arquivados
	archivedCount, err := r.collection.CountDocuments(ctx, bson.M{
		"user_id":     userID,
		"is_archived": true,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao contar arquivados: %w", err)
	}
	stats.Archived = archivedCount

	// Conta atrasados
	overdueCount, err := r.collection.CountDocuments(ctx, bson.M{
		"user_id":  userID,
		"due_date": bson.M{"$lt": time.Now()},
		"status":   bson.M{"$ne": enums.StatusCompleted},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao contar atrasados: %w", err)
	}
	stats.Overdue = overdueCount

	return stats, nil
}

// GetOverdueTodos busca todos atrasados
func (r *todoRepository) GetOverdueTodos(ctx context.Context, userID primitive.ObjectID) ([]*entities.Task, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	filter := bson.M{
		"user_id":     userID,
		"due_date":    bson.M{"$lt": time.Now()},
		"status":      bson.M{"$ne": enums.StatusCompleted},
		"is_archived": false,
	}

	opts := options.Find().SetSort(bson.M{"due_date": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar todos atrasados: %w", err)
	}
	defer cursor.Close(ctx)

	var todos []*entities.Task
	for cursor.Next(ctx) {
		var todo entities.Task
		if err := cursor.Decode(&todo); err != nil {
			return nil, fmt.Errorf("erro ao decodificar todo: %w", err)
		}
		todos = append(todos, &todo)
	}

	return todos, nil
}
