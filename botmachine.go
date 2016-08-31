package botmeans

import (
	"log"
	"time"
)

type Update interface {
	ChatId() int64
	ErrorChan() chan error
}
type Session interface{}
type Result interface{}

type BotMachine struct {
	Logger log.Logger
}

type Handler interface {
	Handle(Update)
}

//из апдейта телеграма извлекаем сессию. вней — понятая инфа из апдейта. она умеет отдавать список команд на выполнение
func (machine BotMachine) RunMachine(updates chan Update, handler Handler, stopChan chan interface{}) {
	updatesChanMap := make(map[int64]chan Update)
	handlerClosedChan := make(chan int64)

	chatFunc := func(ch chan Update, chatID int64) {
		defer func() {
			if r := recover(); r != nil {
				machine.Logger.Println(r)
				handlerClosedChan <- chatID
			}
		}()
		exitSignaller := time.After(10 * time.Second)
		for {
			select {
			case update := <-ch:
				handler.Handle(update)
				exitSignaller = time.After(10 * time.Second)
			case <-exitSignaller:
				handlerClosedChan <- chatID
				return
			}
		}
	}
	for {
		select {
		case update := <-updates:
			chatID := update.ChatId()
			var chatChan chan Update
			ok := false
			if chatChan, ok = updatesChanMap[chatID]; !ok {
				chatChan = make(chan Update)
				updatesChanMap[chatID] = chatChan
				go chatFunc(chatChan, chatID)
			}
			go func() { chatChan <- update }()
		case id := <-handlerClosedChan:
			delete(updatesChanMap, id)
		case <-stopChan:
			return
		}
	}

}
