package api

import (
	"chat-demo/connect"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"net/http"
)

type Server struct {
	app *fiber.App
	db  *connect.Wrapper
	ws  *WsHandler
}

type errorResp struct {
	Error string `json:"error"`
}

func newErrorResp(err error) errorResp {
	return errorResp{
		Error: err.Error(),
	}
}

func Create(db *connect.Wrapper) (*Server, error) {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("FIBER")
	})

	srv := &Server{
		app: app,
		db:  db,
		ws:  CreateWsHandler(db),
	}
	go srv.ws.WriteMessageBack()

	app.Use(cors.New(cors.Config{}))
	app.Use(logger.New())

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(conn *websocket.Conn) {
		srv.ws.GenerateConnection(conn.Params("id"), conn)
	}))

	srv.Routes()
	return srv, nil
}

func (s *Server) Listen(port string) error {
	return s.app.Listen(":" + port)
}

func (s *Server) Routes() {
	s.app.Get("/user/get_all", s.UserGetAll)
	s.app.Post("/user/register", s.UserCreate)
	s.app.Get("/user/get", s.Auth, s.UserGet)
	s.app.Get("/user/me", s.Auth, s.UserMe)

	s.app.Post("/chat/create", s.Auth, s.ChatCreate)
	s.app.Get("chat/get_all", s.Auth, s.ChatGetAll)
	s.app.Get("/chat/all_by_user", s.Auth, s.ChatGetByUser)
	s.app.Put("/chat/join", s.Auth, s.ChatJoin)
	s.app.Get("/chat/:id", s.Auth, s.GetChatById)
	s.app.Delete("/chatdelete/:chat_id", s.Auth, s.ChatDelete)

	s.app.Post("/msg/create", s.Auth, s.MessageCreate)
	s.app.Delete("/msg/delete/:msg_id", s.Auth, s.MessageDelete)
	s.app.Put("/msg/update", s.Auth, s.MessageUpdate)
	s.app.Get("/msg/get/:chat_id", s.Auth, s.MessageGetAllByChat)

	s.app.Post("/login", s.Login)

}

func BadRequest(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusBadRequest).JSON(newErrorResp(err))
}

func InternalError(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusInternalServerError).JSON(newErrorResp(err))
}

func JsonOk(c *fiber.Ctx, intfs interface{}) error {
	return c.Status(http.StatusOK).JSON(intfs)
}
func Unauthorized(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusUnauthorized).JSON(newErrorResp(err))
}
