package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID   primitive.ObjectID `bson:"_id"`
	GUID string             `bson:"guid"`
}
