package api

import (
	"chat-demo/models"
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"time"
)

func (s *Server) Auth(c *fiber.Ctx) error {

	auth := c.Get("X-Auth-Token")

	token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		return Unauthorized(c, err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return Unauthorized(c, errors.New("Error expiration date check!"))
		}
		u := models.User{}

		str, err := primitive.ObjectIDFromHex(claims["id"].(string))
		if err != nil {
			return Unauthorized(c, err)
		}
		err = s.db.Users().FindOne(context.TODO(), bson.M{"_id": str}).Decode(&u)
		if err != nil {
			return Unauthorized(c, err)
		}

		c.Locals("user_id", u.Id.Hex())

		return c.Next()
	} else {
		return Unauthorized(c, errors.New("Error parsing token"))
	}

}

