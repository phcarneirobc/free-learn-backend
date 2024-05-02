package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dbConfig struct {
	Dbname  string
	Client  *mongo.Client
	Context context.Context
}

var Instance *dbConfig

func StartDB() error {
	uri := getDbConnectionString()
	client, ctx, _, err := connect(uri)
	if err != nil {
		return err
	}

	Instance = &dbConfig{
		Dbname:  getDbName(),
		Client:  client,
		Context: ctx,
	}

	return nil
}

func connect(
	uri string,
) (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

func Close() error {
	return close(Instance.Client, Instance.Context)
}

func close(client *mongo.Client, ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	return client.Disconnect(ctx)
}

func InsertOne(
	client *mongo.Client,
	ctx context.Context,
	dataBase, col string,
	doc interface{},
) (*mongo.InsertOneResult, error) {

	// select database and collection ith Client.Database method
	// and Database.Collection method
	collection := client.Database(dataBase).Collection(col)

	// InsertOne accept two argument of type Context
	// and of empty interface
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}
