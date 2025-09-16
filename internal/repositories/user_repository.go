package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/devgugga/todo-it/internal/database"
	"github.com/devgugga/todo-it/internal/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository interface define os métodos do repositório de usuários
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, page, limit int64) ([]*entities.User, int64, error)
	Exists(ctx context.Context, email string) (bool, error)
}

// userRepository implementa UserRepository
type userRepository struct {
	collection *mongo.Collection
}

// NewUserRepository cria uma nova instância do repositório
func NewUserRepository(db database.Client) UserRepository {
	collections := database.GetCollections(db)

	return &userRepository{
		collection: collections.Users,
	}
}

// Create cria um novo usuário
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	// Prepara entidade para criação
	user.PrepareForCreate()

	// Insere no banco
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("usuário com este email já existe")
		}
		return fmt.Errorf("erro ao criar usuário: %w", err)
	}

	// Atualiza o ID na entidade
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}

	return nil
}

// GetByID busca usuário por ID
func (r *userRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	var user entities.User
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("usuário não encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return &user, nil
}

// GetByEmail busca usuário por email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	var user entities.User
	filter := bson.M{"email": email}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("usuário não encontrado")
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return &user, nil
}

// Update atualiza um usuário
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	// Prepara entidade para atualização
	user.PrepareForUpdate()

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"avatar":     user.Avatar,
			"is_active":  user.IsActive,
			"updated_at": user.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao deletar usuário: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("usuário não encontrado")
	}

	return nil
}

// List lista usuários com paginação
func (r *userRepository) List(ctx context.Context, page, limit int64) ([]*entities.User, int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	// Calcula skip
	skip := (page - 1) * limit

	// Filter para usuários ativos
	filter := bson.M{"is_active": true}

	// Conta total de documentos
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao contar usuários: %w", err)
	}

	// Opções de busca
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1}) // Mais recentes primeiro

	// Executa busca
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar usuários: %w", err)
	}
	defer cursor.Close(ctx)

	// Decodifica resultados
	var users []*entities.User
	for cursor.Next(ctx) {
		var user entities.User
		if err := cursor.Decode(&user); err != nil {
			return nil, 0, fmt.Errorf("erro ao decodificar usuário: %w", err)
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, fmt.Errorf("erro no cursor: %w", err)
	}

	return users, total, nil
}

// Exists verifica se um usuário com o email existe
func (r *userRepository) Exists(ctx context.Context, email string) (bool, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
	}

	filter := bson.M{"email": email}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("erro ao verificar existência do usuário: %w", err)
	}

	return count > 0, nil
}

// Delete remove um usuário (soft delete)
func (r *userRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("erro ao deletar usuário: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("usuário não encontrado")
	}

	return nil
}
