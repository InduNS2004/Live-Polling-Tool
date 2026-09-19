package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Mongo struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func Connect(ctx context.Context, uri, dbName string) (*Mongo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err = client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}
	return &Mongo{Client: client, DB: client.Database(dbName)}, nil
}
func (m *Mongo) Close(ctx context.Context) error { return m.Client.Disconnect(ctx) }
