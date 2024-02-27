package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type RefreshSession struct {
	GUID         string             `bson:"guid"`
	RefreshToken string             `bson:"refresh_token"`
	ExpiresIn    primitive.DateTime `bson:"expires_in"`
}
