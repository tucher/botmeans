package botmeans

import (
	"time"
)

type ChatIdentifier interface {
	ChatId() int64
}

type Executer interface {
	Id() int64
	Execute()
}

func RunMachine(queueStream chan Executer, interval time.Duration) chan interface{} {
	stopChan := make(chan interface{})
	queueChanMap := make(map[int64]chan Executer)
	handlerClosedChan := make(chan int64)

	handler := func(ch chan Executer, ID int64) {
		defer func() {
			if r := recover(); r != nil {
				handlerClosedChan <- ID
			}
		}()
		exitSignaller := time.After(interval)
		for {
			select {
			case queue := <-ch:
				queue.Execute()
				exitSignaller = time.After(interval)
			case <-exitSignaller:
				handlerClosedChan <- ID
				return
			}
		}
	}
	go func() {
		for {
			select {
			case queue := <-queueStream:
				if queue == nil {
					continue
				}
				ID := queue.Id()
				var queueChan chan Executer
				ok := false
				if queueChan, ok = queueChanMap[ID]; !ok {
					queueChan = make(chan Executer)
					queueChanMap[ID] = queueChan
					go handler(queueChan, ID)
				}
				// go func() { queueChan <- queue }()
				queueChan <- queue
			case id := <-handlerClosedChan:
				delete(queueChanMap, id)
			case <-stopChan:
				return
			}
		}
	}()
	return stopChan
}
