// Package botmeans provides a framework for creation of complex high-loaded Telegram bots with rich behaviour
package botmeans

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//MeansBot is a body of botmeans framework instance.
type MeansBot struct {
	bot        *tgbotapi.BotAPI
	db         *gorm.DB
	Logger     *log.Logger
	telegramID int64
	BotName    string
	netConfig  NetConfig
	tlgConfig  TelegramConfig

	stopChan chan interface{}
}

//NetConfig is a MeansBot network config for using with New function
type NetConfig struct {
	ListenIP   string
	ListenPort int16
}

//TelegramConfig is a MeansBot telegram API config for using with New function
type TelegramConfig struct {
	BotToken    string
	WebhookHost string
	SSLCertFile string
}

//New creates new MeansBot instance
func New(DB *gorm.DB, netConfig NetConfig, tlgConfig TelegramConfig) (*MeansBot, error) {
	if DB == nil {
		return &MeansBot{}, fmt.Errorf("No db connection given")
	}
	bot, err := tgbotapi.NewBotAPI(tlgConfig.BotToken)
	if err != nil {
		return &MeansBot{}, err
	}

	botID, _ := strconv.ParseInt(strings.Split(tlgConfig.BotToken, ":")[0], 10, 64)

	ret := &MeansBot{
		bot:        bot,
		db:         DB,
		Logger:     log.New(os.Stdout, "MeansBot: ", log.Llongfile|log.Ldate|log.Ltime),
		telegramID: botID,
		netConfig:  netConfig,
		tlgConfig:  tlgConfig,
		stopChan:   make(chan interface{}),
	}

	ret.bot.RemoveWebhook()
	_, err = ret.bot.SetWebhook(tgbotapi.NewWebhookWithCert(fmt.Sprintf("https://%v:8443/%v", ret.tlgConfig.WebhookHost, ret.bot.Token),
		ret.tlgConfig.SSLCertFile))
	if err != nil {
		ret.Logger.Println("Failed to set webhook", err)
	}
	SessionInitDB(DB)
	BotMessageInitDB(DB)

	return ret, nil
}

//Stop function stops the bot, which was started by Run
func (ui *MeansBot) Stop() {
	ui.stopChan <- struct{}{}
}

//Run starts updates handling
func (ui *MeansBot) Run(handlersProvider ActionHandlersProvider, templateDir string) error {

	actionFactory := func(
		session SessionInterface,
		cmdGetter func() string,
		argsGetter func() []ArgInterface,
		sourceMessageGetter func() BotMessageInterface,
		out chan Executer,
	) {
		ActionFactory(
			session,
			cmdGetter,
			argsGetter,
			sourceMessageGetter,
			&Sender{
				session:     session,
				bot:         ui.bot,
				templ:       getTemplater(),
				templateDir: templateDir,
				msgFactory:  func() BotMessageInterface { return NewBotMessage(session.ChatId(), ui.db) },
			},
			out,
			handlersProvider,
		)
	}

	sessionFactory := func(base SessionBase) (SessionInterface, error) {
		return SessionLoader(base, ui.db, ui.BotName, ui.telegramID, ui.bot)
	}

	aliaser := AliaserFromTemplateDir(templateDir)

	argsParser := func(tgUpdate tgbotapi.Update) []ArgInterface {
		return ArgsParser(tgUpdate, sessionFactory, aliaser)
	}

	cmdParser := func(tgUpdate tgbotapi.Update) string {
		return CmdParser(tgUpdate, aliaser)
	}

	botMsgFactory := func(chatID int64, msgId int64, callbackID string) BotMessageInterface {
		return BotMessageDBLoader(chatID, msgId, callbackID, ui.db)
	}

	updatesChan := ui.bot.ListenForWebhook("/" + ui.bot.Token)
	go http.ListenAndServe(fmt.Sprintf("%v:%v", ui.netConfig.ListenIP, ui.netConfig.ListenPort), nil)

	actionsChan := createTGUpdatesParser(
		updatesChan,
		sessionFactory,
		actionFactory,
		botMsgFactory,
		cmdParser,
		argsParser,
	)

	ui.stopChan = RunMachine(actionsChan, time.Minute)
	return nil
}
