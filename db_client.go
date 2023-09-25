package hin

import (
	"context"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func NewMongoDB(logger *Logger) *mongo.Client {
	if !viper.IsSet("mongo.uri") {
		logger.Warn("WARN: mongo connect info was use for local test")
		viper.Set("mongo.uri", "mongodb://localhost:27017/")
	}

	opts := options.Client().ApplyURI(viper.GetString("mongo.uri"))

	if viper.GetBool("server.debug") && viper.GetBool("mongo.log") {
		opts.SetMonitor(&event.CommandMonitor{
			Started: func(_ context.Context, evt *event.CommandStartedEvent) {
				log.Println(evt.Command)
			},
		})
	}

	if client, err := mongo.Connect(context.Background(), opts); err != nil {
		panic(err)
	} else {
		return client
	}
}
