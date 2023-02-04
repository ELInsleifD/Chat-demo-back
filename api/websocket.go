package api

import (
	"chat-demo/connect"
	"chat-demo/models"
	"context"
	"encoding/json"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)

type WsHandler struct {
	wsArrMap   map[string]map[string]*websocket.Conn
	StreamMsgs chan ToSend
	datab      *connect.Wrapper
}
type ToSend struct {
	event   string
	data    string
	userids []string
}

var (
	msg []byte
	err error
)

func (ws *WsHandler) GenerateConnection(id string, v *websocket.Conn) {
	conid := primitive.NewObjectID().Hex()
	if ws.wsArrMap[id] == nil {
		ws.wsArrMap[id] = map[string]*websocket.Conn{}
	}
	ws.wsArrMap[id][conid] = v

	for {
		_, msg, err = v.ReadMessage()
		if err != nil {
			delete(ws.wsArrMap[id], conid)
			break
		}
		var result ToSend
		err := json.Unmarshal(msg, &result)
		if err != nil {
			log.Println("error unmarshaling result in websocket")
			continue
		}

		ws.StreamMsgs <- result

		log.Printf("recv: %s", msg)

	}

}

func (ws *WsHandler) WriteMessageBack() {
	for {
		mock := <-ws.StreamMsgs

		type msg struct {
			Event string `json:"event"`
			Data  string `json:"data"`
		}
		msgg := msg{
			Event: mock.event,
			Data:  mock.data,
		}
		mybyte, err := json.Marshal(msgg)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, elem := range mock.userids {

			_, ok := (ws.wsArrMap)[elem]
			if !ok {
				continue
			}

			for _, ele := range (ws.wsArrMap)[elem] {

				err := ele.WriteMessage(1, mybyte)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func (ws *WsHandler) MessageSave(msg ToSend) {

	ws.StreamMsgs <- msg

}

func CreateWsHandler(db *connect.Wrapper) *WsHandler {
	return &WsHandler{
		wsArrMap:   map[string]map[string]*websocket.Conn{},
		StreamMsgs: make(chan ToSend),
		datab:      db,
	}
}

func (ws *WsHandler) eventSort(event string, scoop interface{}, chatid string) {

	hexid, err := primitive.ObjectIDFromHex(chatid)
	if err != nil {
		log.Println("failed to hex")
		return
	}
	result := models.Chat{}
	err = ws.datab.Chats().FindOne(context.TODO(), bson.M{"_id": hexid}).Decode(&result)
	if err != nil {
		log.Println("filed to find chat")
		return
	}

	b, err := json.Marshal(scoop)
	if err != nil {
		log.Println("failed to marshal scoop")
		return
	}

	msg := ToSend{
		event:   event,
		data:    string(b),
		userids: result.UserIds,
	}

	ws.MessageSave(msg)
}

func (ws *WsHandler) MessageCreate(msg message, chatid string) {
	ws.eventSort("msgcreate", msg, chatid)
}
func (ws *WsHandler) MessageDelete(msg message, chatid string) {
	ws.eventSort("msgdelete", msg, chatid)
}
func (ws *WsHandler) MessageUpdate(msg message, chatid string) {
	ws.eventSort("msgupdate", msg, chatid)
}
func (ws *WsHandler) ChannelDelete(chatid string, userids []string) {
	ws.specificChatDelete("chdelete", chatid, userids)
}
func (ws *WsHandler) specificChatDelete(event string, id string, userids []string) {
	msg := ToSend{
		event:   event,
		data:    id,
		userids: userids,
	}

	ws.MessageSave(msg)
}
