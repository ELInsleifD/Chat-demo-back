package api

import (
	"chat-demo/models"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type chat struct {
	Id       string   `json:"id"`
	ChatName string   `json:"chat_name"`
	UserIds  []string `json:"user_ids"`
}

func (s *Server) ChatCreate(c *fiber.Ctx) error {
	ch := chat{}
	err := c.BodyParser(&ch)
	if err != nil {
		return BadRequest(c, err)
	}

	newId := primitive.NewObjectID()

	userId := c.Locals("user_id")

	count, err := s.db.Chats().CountDocuments(context.TODO(), bson.M{"chat_name": ch.ChatName})
	if err != nil {
		return InternalError(c, err)
	}
	if count > 0 {
		return InternalError(c, errors.New("Chat already exists"))
	}

	_, err = s.db.Chats().InsertOne(context.TODO(), models.Chat{
		Id:       newId,
		ChatName: ch.ChatName,
		UserIds:  []string{userId.(string)},
	})
	if err != nil {
		return InternalError(c, err)
	}

	return JsonOk(c, chat{
		Id:       newId.Hex(),
		ChatName: ch.ChatName,
	})
}

func (s *Server) ChatGetAll(c *fiber.Ctx) error {

	cursor, err := s.db.Chats().Find(context.TODO(), bson.M{})
	if err != nil {
		return InternalError(c, err)
	}

	var results []models.Chat
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		return InternalError(c, err)
	}

	jsonChats := make([]chat, len(results))
	for i, elem := range results {
		jsonChats[i] = chat{
			Id:       elem.Id.Hex(),
			ChatName: elem.ChatName,
		}
	}
	return JsonOk(c, jsonChats)
}

func (s *Server) GetChatById(c *fiber.Ctx) error {
	hexid, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return BadRequest(c, err)
	}

	var result chat
	err = s.db.Chats().FindOne(context.TODO(), bson.M{"_id": hexid}).Decode(&result)
	if err != nil {
		return InternalError(c, err)
	}

	return JsonOk(c, result)
}
func (s *Server) ChatGetByUser(c *fiber.Ctx) error {

	cursor, err := s.db.Chats().Find(context.TODO(), bson.M{"user_ids": c.Locals("user_id").(string)})
	if err != nil {
		return InternalError(c, err)
	}

	var results []models.Chat
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		return InternalError(c, err)
	}

	jsonChats := make([]chat, len(results))
	for i, elem := range results {
		jsonChats[i] = chat{
			Id:       elem.Id.Hex(),
			ChatName: elem.ChatName,
			UserIds:  elem.UserIds,
		}
	}
	return JsonOk(c, jsonChats)
}

func (s *Server) ChatJoin(c *fiber.Ctx) error {
	var ch struct {
		ChatName string `json:"chat_name"`
	}
	err := c.BodyParser(&ch)
	if err != nil {
		return BadRequest(c, err)
	}

	var result models.Chat
	err = s.db.Chats().FindOne(context.TODO(), bson.M{"chat_name": ch.ChatName}).Decode(&result)
	if err != nil {
		return InternalError(c, err)
	}
	newUserIdArr := append(result.UserIds, c.Locals("user_id").(string))

	insert, err := s.db.Chats().UpdateOne(context.TODO(), bson.M{"_id": result.Id},
		bson.M{"$set": bson.M{"user_ids": newUserIdArr}})
	if err != nil {
		return InternalError(c, err)
	}
	return JsonOk(c, insert.UpsertedCount)
}
func (s *Server) ChatDelete(c *fiber.Ctx) error {
	hexid, err := primitive.ObjectIDFromHex(c.Params("chat_id"))
	if err != nil {
		return BadRequest(c, err)
	}

	var result models.Chat
	err = s.db.Chats().FindOne(context.TODO(), bson.M{"_id": hexid}).Decode(&result)
	if err != nil {
		return InternalError(c, err)
	}

	_, err = s.db.Chats().DeleteOne(context.TODO(), bson.M{"_id": hexid})
	if err != nil {
		return InternalError(c, err)
	}
	_, err = s.db.Messages().DeleteMany(context.TODO(), bson.M{"chat_id": c.Params("chat_id")})
	if err != nil {
		return InternalError(c, err)
	}

	s.ws.ChannelDelete(hexid.Hex(), result.UserIds)

	return c.Status(http.StatusOK).SendString("deleted" + hexid.Hex())
}
