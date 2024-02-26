package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type RefreshSession struct {
	ID           primitive.ObjectID `bson:"_id"`
	RefreshToken string             `bson:"refresh_token"`
}
