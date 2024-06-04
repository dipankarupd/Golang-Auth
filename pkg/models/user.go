package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id           primitive.ObjectID `bson:"_id"`
	Name         *string            `json:"name,omitempty" validate:"required,min=2,max=100"`
	Email        *string            `json:"email" validate:"email,required"`
	Password     *string            `json:"password" validate:"required,min=6"`
	AccessToken  *string            `json:"access_token"`
	RefreshToken *string            `json:"request_token"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	UserType     *string            `json:"user_type"`
	User_id      string             `json:"user_id"`
}
