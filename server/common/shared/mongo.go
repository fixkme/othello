package shared

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	MongoClient *mongo.Client
)

func InitMongo(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.Client()
	opts.ApplyURI(uri).
		SetReadPreference(readpref.Primary()).
		SetBSONOptions(&options.BSONOptions{
			UseJSONStructTags: true,
			NilMapAsEmpty:     true,
			NilSliceAsEmpty:   true,
		})

	cli, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}
	if err := cli.Ping(ctx, nil); err != nil {
		return err
	}
	MongoClient = cli
	return nil
}
