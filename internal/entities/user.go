package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Avatar    string             `bson:"avatar,omitempty"`
	IsActive  bool               `bson:"is_active"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (u *User) PrepareForCreate() {
	now := time.Now()
	u.ID = primitive.NewObjectID()
	u.CreatedAt = now
	u.UpdatedAt = now
	u.IsActive = true
}

func (u *User) PrepareForUpdate() {
	u.UpdatedAt = time.Now()
}

func (u *User) GetCollectionName() string {
	return "users"
}
