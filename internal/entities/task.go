package entities

import (
	"time"

	"github.com/devgugga/todo-it/internal/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id"`
	Title       string             `bson:"title"`
	Description string             `bson:"description,omitempty"`
	Status      enums.TaskStatus   `bson:"status"`
	Priority    enums.TaskPriority `bson:"priority"`
	DueDate     *time.Time         `bson:"due_date,omitempty"`
	Tags        []string           `bson:"tags,omitempty"`
	IsArchived  bool               `bson:"is_archived"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty"`
}

func (t *Task) PrepareForCreate(userID primitive.ObjectID) {
	now := time.Now()
	t.ID = primitive.NewObjectID()
	t.UserID = userID
	t.CreatedAt = now
	t.UpdatedAt = now
	t.IsArchived = false

	if t.Status == "" {
		t.Status = enums.StatusPending
	}

	if t.Priority == "" {
		t.Priority = enums.PriorityMedium
	}
}

func (t *Task) PrepareForUpdate() {
	t.UpdatedAt = time.Now()
}

func (t *Task) MarkAsCompleted() {
	now := time.Now()
	t.Status = enums.StatusCompleted
	t.CompletedAt = &now
	t.UpdatedAt = now
}

func (t *Task) MarkAsPending() {
	t.Status = enums.StatusPending
	t.CompletedAt = nil
	t.UpdatedAt = time.Now()
}

func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == enums.StatusCompleted {
		return false
	}
	return t.DueDate.Before(time.Now())
}

func (t *Task) GetCollectionName() string {
	return "tasks"
}
