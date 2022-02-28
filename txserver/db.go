package main

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ()

func updateUserAccount(ctx *context.Context, username string, update primitive.M) error {
	filter := bson.M{"username": username}

	accountsCollection := client.Database("test").Collection("Accounts")
	result, err := accountsCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf("Error updating account with username: %s, error: %s", username, err.Error())
		return errors.New("account update unsuccessful")
	}

	if result.MatchedCount == 1 {
		return nil
	}

	return errors.New("account update unsuccessful")
}
