package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Chat struct {
	Id       primitive.ObjectID `bson:"_id"`
	ChatName string             `bson:"chat_name"`
	UserIds  []string           `bson:"user_ids"`
}
