package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Message struct {
	Id       primitive.ObjectID `bson:"_id"`
	Content  string             `bson:"content,omitempty"`
	UserId   string             `bson:"user_id"`
	ChatId   string             `bson:"chat_id"`
	UserName string             `bson:"user_name"`
}
