// Package botmeans provides a framework for creation of complex high-loaded Telegram bots with rich behaviour
package botmeans

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type MessengerAdapter interface {
	Run(chan Executer)
	Stop()
	Init(DB *gorm.DB)
	IsOneToOne(s *Session) bool
}

//MeansBot is a body of botmeans framework instance.
type MeansBot struct {
	db       *gorm.DB
	adapters []MessengerAdapter
}

//New creates new MeansBot instance
func New(DB *gorm.DB, adapters []MessengerAdapter) (*MeansBot, error) {
	if DB == nil {
		return &MeansBot{}, fmt.Errorf("No db connection given")
	}

	ret := &MeansBot{
		db:       DB,
		adapters: adapters,
	}

	for _, adapter := range adapters {
		adapter.Init(DB)
	}
	SessionInitDB(DB)
	BotMessageInitDB(DB)
	UserInitDB(DB)
	return ret, nil
}

//Run starts updates handling. Returns stop chan
func (this *MeansBot) Run(handlersProvider ActionHandlersProvider) chan interface{} {
	// templateDir := ui.tlgConfig.TemplateDir

	actionsChan := make(chan Executer)

	for _, adapter := range this.adapters {
		go adapter.Run(actionsChan)
	}

	return RunMachine(actionsChan, time.Minute)

	// actionFactory := func(
	// 	sessionBase SessionBase,
	// 	sessionFactory SessionFactory,
	// 	getters actionExecuterFactoryConfig,
	// 	out chan Executer,
	// ) {
	// 	ActionFactory(
	// 		sessionBase,
	// 		sessionFactory,
	// 		getters,
	// 		func(s senderSession) SenderInterface {
	// 			return &Sender{
	// 				session:     s,
	// 				bot:         ui.bot,
	// 				templateDir: templateDir,
	// 				msgFactory:  func() BotMessageInterface { return NewBotMessage(s.ChatId(), ui.db) },
	// 			}
	// 		},
	// 		out,
	// 		handlersProvider,
	// 	)
	// }

	// aliaser := AliaserFromTemplateDir(templateDir)
	// argsParser := func(tgUpdate tgbotapi.Update) Args {
	// 	return ArgsParser(tgUpdate, sessionFactory, aliaser)
	// }
	// cmdParser := func(tgUpdate tgbotapi.Update) string {
	// 	return CmdParser(tgUpdate, aliaser)
	// }
	// botMsgFactory := func(chatID int64, msgId int64, callbackID string) BotMessageInterface {
	// 	return BotMessageDBLoader(chatID, msgId, callbackID, ui.db)
	// }

}
