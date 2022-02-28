package main

import (
	"sync"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/sets/hashset"
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

	c1 := a.(float32)
	c2 := b.(float32)

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

func (p *poll) buy_poll(list *map[string]*treemap.Map, lock *sync.Mutex) []byte {

	if p.buyPollingInProgress {
		return nil
	}

	go polling_thread(list, lock, &p.buyPollingInProgress)
	p.buyPollingInProgress = true
	return nil
}

func (p *poll) sell_poll(list *map[string]*treemap.Map, lock *sync.Mutex) []byte {

	if p.sellPollingInProgress {
		return nil
	}

	go polling_thread(list, lock, &p.sellPollingInProgress)
	p.sellPollingInProgress = true
	return nil
}

func polling_thread(list *map[string]*treemap.Map, lock *sync.Mutex, pollingInProgress *bool) {

	for {
		break
	}
}

func trigger(command *Command, trigger string) []byte {

	var lock *sync.Mutex
	var list *map[string]*treemap.Map
	var trigger_poll func(list *map[string]*treemap.Map, lock *sync.Mutex) []byte

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

	price_wait_list, found := (*list)[command.Stock]
	if !found {
		user_list := hashset.New()
		price_wait_list := treemap.NewWith(float32Comparator)

		user_list.Add(command.Username)
		price_wait_list.Put(command.Amount, user_list)
		(*list)[command.Stock] = price_wait_list

		return trigger_poll(list, lock)
	}

	Iuser_list, found := price_wait_list.Get(command.Amount)
	user_list := Iuser_list.(*hashset.Set)
	if !found {
		user_list := hashset.New()
		user_list.Add(command.Username)
		price_wait_list.Put(command.Amount, user_list)
		(*list)[command.Stock] = price_wait_list

		return trigger_poll(list, lock)
	}

	user_list.Add(command.Username)
	price_wait_list.Put(command.Amount, user_list)
	(*list)[command.Stock] = price_wait_list

	return trigger_poll(list, lock)
}
