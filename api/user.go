package api

import (
	"chat-demo/models"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

type user struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret,omitempty"`
}

func (s *Server) UserCreate(c *fiber.Ctx) error {

	u := user{}
	if err := c.BodyParser(&u); err != nil {
		return BadRequest(c, err)
	}

	count, err := s.db.Users().CountDocuments(context.TODO(), bson.M{"name": u.Name})
	if err != nil {
		return InternalError(c, err)
	}
	if count > 0 {
		return InternalError(c, errors.New("user already exists"))
	}

	hashedSecret, err := models.HashPass(u.Secret)
	if err != nil {
		return BadRequest(c, err)
	}

	newId := primitive.NewObjectID()

	_, err = s.db.Users().InsertOne(
		context.TODO(), models.User{
			Id:     newId,
			Name:   u.Name,
			Secret: hashedSecret,
		},
	)
	if err != nil {
		return InternalError(c, err)
	}

	return c.Status(http.StatusOK).JSON(user{
		Id:   newId.Hex(),
		Name: u.Name,
	})
}

func (s *Server) UserGetAll(c *fiber.Ctx) error {
	cursor, err := s.db.Users().Find(context.TODO(), bson.M{})
	if err != nil {
		return InternalError(c, err)
	}

	var results []models.User
	if err := cursor.All(context.TODO(), &results); err != nil {
		return InternalError(c, err)
	}

	jsonUsers := make([]user, len(results))
	for i, elem := range results {
		jsonUsers[i] = user{
			Id:   elem.Id.Hex(),
			Name: elem.Name,
		}
	}
	return JsonOk(c, jsonUsers)
}

func (s *Server) UserGet(c *fiber.Ctx) error {
	var name string
	err := c.BodyParser(&name)
	if err != nil {
		return BadRequest(c, err)
	}

	var result user
	err = s.db.Users().FindOne(context.TODO(), bson.M{"name": name}).Decode(&result)
	if err != nil {
		return InternalError(c, err)
	}

	return JsonOk(c, result)
}
func (s *Server) UserMe(c *fiber.Ctx) error {
	var result models.User

	hexid, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return InternalError(c, err)
	}

	err = s.db.Users().FindOne(context.TODO(), bson.M{"_id": hexid}).Decode(&result)
	if err != nil {
		return InternalError(c, err)
	}
	return JsonOk(c, user{
		Id:   result.Id.Hex(),
		Name: result.Name,
	})
}
func (s *Server) Login(c *fiber.Ctx) error {
	body := user{}
	err := c.BodyParser(&body)
	if err != nil {
		return BadRequest(c, err)
	}

	var result models.User
	err = s.db.Users().FindOne(context.TODO(), bson.M{"name": body.Name}).Decode(&result)
	if err != nil {
		return InternalError(c, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Secret), []byte(body.Secret))
	if err != nil {
		return BadRequest(c, err)
	}

	claims := jwt.MapClaims{
		"id":  result.Id.Hex(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		return InternalError(c, err)
	}
	return JsonOk(c, fiber.Map{"token": tokenStr})
}
