package api

import (
	"chat-demo/models"
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type message struct {
	Id       string `json:"id"`
	Content  string `json:"content"`
	UserId   string `json:"user_id"`
	ChatId   string `json:"chat_id"`
	UserName string `json:"user_name"`
}

func (s *Server) MessageCreate(c *fiber.Ctx) error {
	msg := message{}
	err := c.BodyParser(&msg)
	if err != nil {
		return BadRequest(c, err)
	}

	newId := primitive.NewObjectID()

	if err != nil {
		return InternalError(c, err)
	}

	hexid, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return InternalError(c, err)
	}

	var userResult models.User
	err = s.db.Users().FindOne(context.TODO(), bson.M{"_id": hexid}).Decode(&userResult)
	if err != nil {
		return InternalError(c, err)
	}

	_, err = s.db.Messages().InsertOne(context.TODO(), models.Message{
		Id:       newId,
		Content:  msg.Content,
		UserId:   c.Locals("user_id").(string),
		ChatId:   msg.ChatId,
		UserName: userResult.Name,
	})
	if err != nil {
		return InternalError(c, err)
	}
	msg = message{
		Id:       newId.Hex(),
		Content:  msg.Content,
		UserId:   c.Locals("user_id").(string),
		ChatId:   msg.ChatId,
		UserName: userResult.Name,
	}

	s.ws.MessageCreate(msg, msg.ChatId)

	return JsonOk(c, msg)
}

func (s *Server) MessageGetAllByChat(c *fiber.Ctx) error {

	cursor, err := s.db.Messages().Find(context.TODO(), bson.M{"chat_id": c.Params("chat_id")})
	if err != nil {
		return InternalError(c, err)
	}

	var results []models.Message
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		return InternalError(c, err)
	}

	jsonMsgs := make([]message, len(results))
	for i, elem := range results {
		jsonMsgs[i] = message{
			Id:       elem.Id.Hex(),
			Content:  elem.Content,
			UserId:   elem.UserId,
			ChatId:   elem.ChatId,
			UserName: elem.UserName,
		}
	}
	return JsonOk(c, jsonMsgs)
}

func (s *Server) MessageUpdate(c *fiber.Ctx) error {

	newMsg := message{}
	err := c.BodyParser(&newMsg)
	if err != nil {
		return BadRequest(c, err)
	}

	hexid, err := primitive.ObjectIDFromHex(newMsg.Id)

	_, err = s.db.Messages().UpdateOne(context.TODO(), bson.M{"_id": hexid}, bson.M{
		"$set": bson.M{"content": newMsg.Content}})
	if err != nil {
		return InternalError(c, err)
	}

	s.ws.MessageUpdate(newMsg, newMsg.ChatId)

	return JsonOk(c, newMsg.Content)
}

func (s *Server) MessageDelete(c *fiber.Ctx) error {
	hexid, err := primitive.ObjectIDFromHex(c.Params("msg_id"))
	if err != nil {
		return BadRequest(c, err)
	}

	var result models.Message
	err = s.db.Messages().FindOne(context.TODO(), bson.M{"_id": hexid}).Decode(&result)
	if err != nil {
		return InternalError(c, err)
	}

	_, err = s.db.Messages().DeleteOne(context.TODO(), bson.M{"_id": hexid})
	if err != nil {
		return InternalError(c, err)
	}

	msg := message{
		Id:       hexid.Hex(),
		Content:  result.Content,
		UserId:   result.UserId,
		ChatId:   result.ChatId,
		UserName: result.UserName,
	}

	s.ws.MessageDelete(msg, msg.ChatId)

	return c.Status(http.StatusOK).SendString("deleted msg: " + hexid.Hex())
}
