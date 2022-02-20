package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var collection *mongo.Collection

func addtoDB() {

	doc := bson.D{{"user", "sadfd23sfds"}, {"account_balance", "98742"}}

	result, err := collection.InsertOne(context.TODO(), doc)

	if err != nil {
		panic(err)
	}

	log.Printf("Inserted document with _id: %v\n", result.InsertedID)
}

func setupDB() {

	const uri = "mongodb://mongodb_container:27017/?maxPoolSize=20&w=majority"
	var client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	log.Printf("Successfully connected and pinged.")

	collection = client.Database("test").Collection("transactions")

}
