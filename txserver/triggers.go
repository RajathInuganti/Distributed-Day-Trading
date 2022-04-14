package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/sets/hashset"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
Buy & Sell lists are in the form:
	{
		stock a: {
			price a wait list: user_list
			price b wait list: user_list
		}
		stock b: {
			price a wait list: user_list
			price b wait list: user_list
		}
	}
*/

var (
	ctx              *context.Context
	poller           = new(poll)
	command          *Command
	price_adjustment bool
	previous_price   float64
	buy_list         = make(map[string]*treemap.Map)
	sell_list        = make(map[string]*treemap.Map)
)

func float64Comparator(a, b interface{}) int {

	c1 := a.(float64)
	c2 := b.(float64)

	switch {
	case c1 > c2:
		return -1
	case c1 < c2:
		return 1
	default:
		return 0
	}

}

type poll struct {
	run_buy_polling       bool
	run_sell_polling      bool
	buyPollingInProgress  bool
	sellPollingInProgress bool
	buy_updates           chan bool
	sell_updates          chan bool
}

func (p *poll) buy_poll() []byte {

	if p.buyPollingInProgress {
		poller.run_buy_polling = false
		poller.buy_updates <- true
		return []byte("Buy trigger polling initiated")
	}

	p.buyPollingInProgress = true
	poller.buy_updates = make(chan bool)
	go trigger_polling("BUY")
	poller.buy_updates <- true
	return []byte("Buy trigger polling initiated")
}

func (p *poll) sell_poll() []byte {

	if p.sellPollingInProgress {
		poller.run_sell_polling = false
		poller.sell_updates <- true
		return []byte("Sell trigger polling initiated")
	}

	p.sellPollingInProgress = true
	poller.sell_updates = make(chan bool)
	go trigger_polling("SELL")
	poller.sell_updates <- true
	return []byte("Sell trigger polling initiated")
}

func trigger(context *context.Context, cmd *Command, adjustment bool, price float64, trigger string) []byte {

	ctx = context
	command = cmd
	price_adjustment = adjustment
	previous_price = price

	if trigger == "BUY" {
		return poller.buy_poll()
	}

	return poller.sell_poll()

}

func trigger_polling(trigger string) {

	var run_polling *bool
	var updates chan bool
	var list *map[string]*treemap.Map

	if trigger == "BUY" {
		list = &buy_list
		updates = poller.buy_updates
		run_polling = &poller.run_buy_polling
	} else {
		list = &sell_list
		updates = poller.sell_updates
		run_polling = &poller.run_sell_polling
	}

	for {
		select {
		case <-updates:
			price_wait_list, found := (*list)[command.Stock]
			if !found {
				user_list := hashset.New()
				price_wait_list := treemap.NewWith(float64Comparator)
				user_list.Add(command.Username)
				price_wait_list.Put(command.Amount, user_list)
				(*list)[command.Stock] = price_wait_list

				(*run_polling) = true
				break
			}

			if price_adjustment && previous_price != command.Amount {
				Iprevious_price_user_list, _ := price_wait_list.Get(previous_price)
				previous_price_user_list := Iprevious_price_user_list.(*hashset.Set)
				previous_price_user_list.Remove(command.Username)
				price_wait_list.Put(previous_price, previous_price_user_list)
				if previous_price_user_list.Empty() {
					price_wait_list.Remove(previous_price)
				}
				(*list)[command.Stock] = price_wait_list
			}

			Iuser_list, found := price_wait_list.Get(command.Amount)
			if !found {
				user_list := hashset.New()
				user_list.Add(command.Username)
				price_wait_list.Put(command.Amount, user_list)
				(*list)[command.Stock] = price_wait_list

				(*run_polling) = true
				break
			}

			user_list := Iuser_list.(*hashset.Set)
			user_list.Add(command.Username)
			price_wait_list.Put(command.Amount, user_list)
			(*list)[command.Stock] = price_wait_list

			(*run_polling) = true

		default:
			for stock := range *list {

				if !(*run_polling) {
					break
				}

				quote, err := get_quote(stock, os.Getenv("HOSTNAME"))
				for err != nil {
					quote, err = get_quote(stock, os.Getenv("HOSTNAME"))
				}
				quoted_price, err := strconv.ParseFloat(quote[0], 64)
				if err != nil {
					log.Println("error parsing string")
				}

				price_wait_list := (*list)[stock]
				priceIterator := price_wait_list.Iterator()

				for priceIterator.Next() {

					price := priceIterator.Key().(float64)
					if price < quoted_price {
						break
					}

					Iuser_list, _ := price_wait_list.Get(price)
					user_list := Iuser_list.(*hashset.Set)
					usernames := user_list.Values()
					go update_account(ctx, trigger, stock, usernames)

					price_wait_list.Remove(price)
					(*list)[stock] = price_wait_list
				}

				if price_wait_list.Empty() {
					delete(*list, stock)
				}
			}
		}
	}
}

func update_account(ctx *context.Context, trigger string, stock string, usernames []interface{}) {

	for index := range usernames {
		username := usernames[index].(string)
		var update primitive.M

		account, err := find_account(ctx, username)
		if err != nil {
			log.Printf("No account found for: %s", username)
		}

		if trigger == "BUY" {
			account.Stocks[stock] += account.BuyAmounts[stock]
			delete(account.BuyAmounts, stock)
			delete(account.BuyTriggers, stock)

			update = bson.M{
				"$set": bson.M{
					"buyAmounts":  account.BuyAmounts,
					"buyTriggers": account.BuyTriggers,
					"stocks":      account.Stocks,
				},
			}
		} else {
			account.Balance += account.SellAmounts[stock]
			delete(account.SellAmounts, stock)
			delete(account.SellTriggers, stock)

			update = bson.M{
				"$set": bson.M{
					"balance":      account.Balance,
					"sellAmounts":  account.SellAmounts,
					"sellTriggers": account.SellTriggers,
				},
			}
		}

		err = updateUserAccount(ctx, username, update, account)
		if err != nil {
			log.Printf("Error updating account")
		}

		log.Println("trigger successfully executed")
	}
}
