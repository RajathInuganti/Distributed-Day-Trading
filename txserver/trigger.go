package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"

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
	poller    poll
	buy_lock  sync.Mutex
	sell_lock sync.Mutex
	buy_list  map[string]*treemap.Map
	sell_list map[string]*treemap.Map
)

func float32Comparator(a, b interface{}) int {

	c1 := a.(float64)
	c2 := b.(float64)

	switch {
	case c1 > c2:
		return 1
	case c1 < c2:
		return -1
	default:
		return 0
	}

}

type poll struct {
	buyPollingInProgress  bool
	sellPollingInProgress bool
}

func (p *poll) buy_poll(ctx *context.Context, list *map[string]*treemap.Map, lock *sync.Mutex) []byte {

	if p.buyPollingInProgress {
		return []byte("Buy trigger polling initiated")
	}

	go polling_thread(ctx, list, lock, &p.buyPollingInProgress, "BUY")
	p.buyPollingInProgress = true
	return []byte("Buy trigger polling initiated")
}

func (p *poll) sell_poll(ctx *context.Context, list *map[string]*treemap.Map, lock *sync.Mutex) []byte {

	if p.sellPollingInProgress {
		return []byte("Sell trigger polling initiated")
	}

	go polling_thread(ctx, list, lock, &p.sellPollingInProgress, "SELL")
	p.sellPollingInProgress = true
	return []byte("Sell trigger polling initiated")
}

func polling_thread(ctx *context.Context, list *map[string]*treemap.Map, lock *sync.Mutex, pollingInProgress *bool, trigger string) {

	for len(*list) != 0 {
		for stock := range *list {

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

				lock.Lock()
				price_wait_list.Remove(price)
				(*list)[stock] = price_wait_list
				lock.Unlock()
			}
		}
	}
	*pollingInProgress = false
}

func update_account(ctx *context.Context, trigger string, stock string, usernames []interface{}) {
	for index := range usernames {
		username := usernames[index].(string)
		var update primitive.M

		account, err := find_account(ctx, username)
		if err != nil {
			log.Println("No account found")
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

		err = updateUserAccount(ctx, username, update)
		if err != nil {
			log.Printf("Error updating account\n")
		}
	}
}

func trigger(ctx *context.Context, command *Command, price_adjustment bool, previous_price float64, trigger string) []byte {

	var lock *sync.Mutex
	var list *map[string]*treemap.Map
	var trigger_poll func(ctx *context.Context, list *map[string]*treemap.Map, lock *sync.Mutex) []byte

	if trigger == "BUY" {
		list = &buy_list
		lock = &buy_lock
		trigger_poll = poller.buy_poll
	} else {
		list = &sell_list
		lock = &sell_lock
		trigger_poll = poller.sell_poll
	}

	lock.Lock()
	defer lock.Unlock()

	if price_adjustment {
		price_wait_list := (*list)[command.Stock]
		Iuser_list, _ := price_wait_list.Get(previous_price)
		user_list := Iuser_list.(*hashset.Set)
		user_list.Remove(command.Username)
		price_wait_list.Put(previous_price, user_list)
		// check if price_wait_list is empty

		Iuser_list, found := price_wait_list.Get(command.Amount)
		if !found {
			user_list := hashset.New()
			user_list.Add(command.Username)
			price_wait_list.Put(command.Amount, user_list)
			(*list)[command.Stock] = price_wait_list

			return trigger_poll(ctx, list, lock)
		}

		user_list = Iuser_list.(*hashset.Set)
		user_list.Add(command.Username)
		price_wait_list.Put(command.Amount, user_list)
		(*list)[command.Stock] = price_wait_list

		return trigger_poll(ctx, list, lock)
	}

	price_wait_list, found := (*list)[command.Stock]
	if !found {
		user_list := hashset.New()
		price_wait_list := treemap.NewWith(float32Comparator)

		user_list.Add(command.Username)
		price_wait_list.Put(command.Amount, user_list)
		(*list)[command.Stock] = price_wait_list

		return trigger_poll(ctx, list, lock)
	}

	Iuser_list, found := price_wait_list.Get(command.Amount)
	user_list := Iuser_list.(*hashset.Set)
	if !found {
		user_list := hashset.New()
		user_list.Add(command.Username)
		price_wait_list.Put(command.Amount, user_list)
		(*list)[command.Stock] = price_wait_list

		return trigger_poll(ctx, list, lock)
	}

	user_list.Add(command.Username)
	price_wait_list.Put(command.Amount, user_list)
	(*list)[command.Stock] = price_wait_list

	return trigger_poll(ctx, list, lock)
}
