package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID                 primitive.ObjectID   `bson:"_id"`
	GUID               string               `bson:"guid"`
	Password           string               `bson:"password"`
	RefreshSessionsIDs []primitive.ObjectID `bson:"refresh_sessions_ids"`
}
