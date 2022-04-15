package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ()

func updateUserAccount(ctx *context.Context, username string, update primitive.M, account *UserAccount) error {
	filter := bson.M{"username": username}

	b, err := json.Marshal(account)
	if err != nil {
		panic(err)
	}
	err_redis := rdb.Set(*ctx, username, b, 0).Err()
	if err_redis != nil {
		panic(err_redis)
	}

	accountsCollection := client.Database("test").Collection("Accounts")
	result, err := accountsCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf("Error updating account with username: %s, error: %s", username, err.Error())
		return errors.New("account update unsuccessful")
	}

	if result.MatchedCount == 1 {
		return nil
	}

	if result.MatchedCount == 0 {
		log.Printf("Found %d accounts with username: %s, inserting most recent version of useraccount into mongodb", result.MatchedCount, username)
		bsonBytes, err := bson.Marshal(account)
		if err != nil {
			log.Printf("Error marshalling account with username: %s, error: %s", username, err.Error())
			return nil
		}

		accountsCollection := client.Database("test").Collection("Accounts")
		_, err = accountsCollection.InsertOne(*ctx, bsonBytes)
		if err != nil {
			log.Printf("Error inserting account with username: %s, error: %s", username, err.Error())
			return err
		}
	}

	return nil
}

func find_account(ctx *context.Context, username string) (*UserAccount, error) {
	var account UserAccount

	if parseErrors.usernameEmpty {
		return &account, errors.New("insufficient information")
	}

	val, err := rdb.Get(*ctx, username).Result()
	if err != nil || err == redis.Nil {
		return nil, err
	}

	b := []byte(val)
	var user_account UserAccount
	err = json.Unmarshal(b, &user_account)
	if err != nil {
		panic(err)
	}

	return &user_account, nil
}

func CreateUserAccount(ctx *context.Context, username string) (*UserAccount, error) {

	var balance float64 = 0

	account := &UserAccount{
		Username:     username,
		Balance:      balance,
		Created:      time.Now().Unix(),
		Updated:      time.Now().Unix(),
		BuyAmounts:   map[string]float64{},
		SellAmounts:  map[string]float64{},
		BuyTriggers:  map[string]float64{},
		SellTriggers: map[string]float64{},
		Stocks:       map[string]float64{},
		Transactions: []*Transaction{},
		RecentBuy:    &CommandHistory{},
		RecentSell:   &CommandHistory{},
	}

	bsonBytes, err := bson.Marshal(account)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(account)
	if err != nil {
		panic(err)
	}
	err_redis := rdb.Set(*ctx, username, b, 0).Err()
	if err_redis != nil {
		panic(err_redis)
	}

	accountsCollection := client.Database("test").Collection("Accounts")
	_, err = accountsCollection.InsertOne(*ctx, bsonBytes)
	if err != nil {
		return nil, err
	}

	return account, nil
}
