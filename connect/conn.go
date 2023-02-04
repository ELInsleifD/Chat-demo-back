package connect

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func Connection(dsn string) (*Wrapper, error) {

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(dsn).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	w := &Wrapper{db: client}
	return w, nil
}

type Wrapper struct {
	db *mongo.Client
}

func (w *Wrapper) GetCol(str string) *mongo.Collection {
	return w.db.Database("db1").Collection(str)
}
func (w *Wrapper) Users() *mongo.Collection {
	return w.GetCol("users")
}

func (w *Wrapper) Chats() *mongo.Collection {
	return w.GetCol("chats")
}
func (w *Wrapper) Messages() *mongo.Collection {
	return w.GetCol("msgs")
}
