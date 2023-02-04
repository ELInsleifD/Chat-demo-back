package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id     primitive.ObjectID `bson:"_id"`
	Name   string             `bson:"name"`
	Secret string             `bson:"secret"`
}

func HashPass(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 8)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
