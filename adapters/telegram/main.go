package telegram

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
)

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
	BotName     string
	TemplateDir string
}

type TelegramAdapter struct {
	bot *tgbotapi.BotAPI

	netConfig NetConfig
	tlgConfig TelegramConfig
}

func New(netConfig NetConfig, tlgConfig TelegramConfig) (*TelegramAdapter, error) {
	bot, err := tgbotapi.NewBotAPI(tlgConfig.BotToken)
	if err != nil {
		return nil, err
	}

	ret := &TelegramAdapter{
		bot:       bot,
		netConfig: netConfig,
		tlgConfig: tlgConfig,
	}

	if os.Getenv("BOTMEANS_TELEGRAM_SET_WEBHOOK") == "TRUE" {
		ret.bot.RemoveWebhook()
		webhookConfig := tgbotapi.NewWebhookWithCert(fmt.Sprintf("https://%v:8443/%v", ret.tlgConfig.WebhookHost, ret.bot.Token),
			ret.tlgConfig.SSLCertFile)
		webhookConfig.MaxConnections = 50
		_, err = ret.bot.SetWebhook(webhookConfig)
		if err != nil {
			return nil, err
		}
		info, _ := ret.bot.GetWebhookInfo()
		log.Printf("Webhook set: %+v", info)
	}

	return ret, nil
}

//Session represents the user in chat.
type Session struct {
	TelegramUserName string

	FirstName string
	LastName  string
	ChatName  string
}

//SessionLoader creates the session and loads the data if the session exists
func SessionLoader(base SessionBase, db *gorm.DB, BotID int64, api *tgbotapi.BotAPI) (SessionInterface, error) {
	TelegramUserID := base.TelegramUserID
	TelegramUserName := base.TelegramUserName
	TelegramChatID := base.TelegramChatID
	if TelegramUserID == 0 && TelegramUserName == "" {
		return nil, fmt.Errorf("Invalid session IDs")
	}
	//TODO!
	if TelegramUserID == BotID {
		return nil, fmt.Errorf("Cannot create the session for myself")
	}
	session := &Session{}
	session.db = db
	found := !db.Where("((telegram_user_id=? and telegram_user_id!=0) or (telegram_user_name=? and telegram_user_name!='')) and telegram_chat_id=?", TelegramUserID, TelegramUserName, TelegramChatID).
		First(session).RecordNotFound()
	err := fmt.Errorf("Unknown")
	if api != nil && (!found || session.FirstName == "" && session.LastName == "") {
		if chatMember, err := api.GetChatMember(tgbotapi.ChatConfigWithUser{TelegramChatID, "", int(TelegramUserID)}); err == nil {
			session.FirstName = chatMember.User.FirstName
			session.LastName = chatMember.User.LastName
		}
	}
	if !found {
		session.isNew = true
		session.TelegramChatID = TelegramChatID
		session.TelegramUserID = TelegramUserID
		session.TelegramUserName = TelegramUserName
		session.CreatedAt = time.Now()
		if api != nil {
			if chat, err := api.GetChat(tgbotapi.ChatConfig{ChatID: session.TelegramChatID}); err == nil {
				session.ChatName = chat.Title
			}
		}
		session.UserData = "{}"
		err = nil

	}
	err = nil
	session.hasLeft = base.hasLeft
	session.hasCome = base.hasCome
	session.TelegramUserID = base.TelegramUserID
	session.TelegramUserName = base.TelegramUserName

	return session, err
}

func (this *TelegramAdapter) Run(actionsChan chan Executer) {
	botID, _ := strconv.ParseInt(strings.Split(ui.bot.Token, ":")[0], 10, 64)

	sessionFactory := func(base SessionBase) (SessionInterface, error) {
		return SessionLoader(base, this.db, botID, this.bot)
	}

	var updatesChan tgbotapi.UpdatesChannel
	if ui.tlgConfig.WebhookHost != "" {
		updatesChan = ui.bot.ListenForWebhook("/" + ui.bot.Token)
		go http.ListenAndServe(fmt.Sprintf("%v:%v", ui.netConfig.ListenIP, ui.netConfig.ListenPort), nil)
	} else {
		var err error
		updatesChan, err = ui.bot.GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Limit: 100, Timeout: 10})
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	createTGUpdatesParser(
		actionsChan,
		updatesChan,
		parserConfig{
			sessionFactory,
			actionFactory,
			botMsgFactory,
			cmdParser,
			argsParser,
		},
	)
}
