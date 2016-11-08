//go:generate gen
package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kardianos/osext"
	_ "github.com/lib/pq"
	"github.com/tucher/botmeans"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type PinnedMsg struct {
	ID        int64
	CreatedAt time.Time
	Text      string
	SessionId int64
	From      string
}

func main() {

	if getConfig(&DBSettings) != nil {
		log.Fatal("Cannot read DB settings")
	}

	DSN := fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v",
		DBSettings.DB_login,
		DBSettings.DB_name,
		DBSettings.DB_pass)
	var err error
	DB, err := gorm.Open("postgres", DSN)
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}

	uiConfig := Config{
		DB:     DB,
		Logger: log.New(os.Stdout, "TelegramUI: ", log.Llongfile|log.Ldate|log.Ltime),
	}
	if getConfig(&uiConfig) != nil {
		log.Fatal("Cannot read UI settings")
	}

	meansBot, err := botmeans.New(DB,
		botmeans.NetConfig{uiConfig.ListenIP, uiConfig.ListenPort},
		botmeans.TelegramConfig{
			BotToken:    uiConfig.BotToken,
			WebhookHost: uiConfig.WebhookHost,
			SSLCertFile: "./MESSAGE_PIN_BOT_PUBLIC.pem",
			TemplateDir: "./templates",
		},
	)

	DB.AutoMigrate(&PinnedMsg{})

	meansBot.Run(
		func(id string) (f botmeans.ActionHandler, s bool) {
			s = true
			switch id {
			case "":
				f = func(c botmeans.ActionContextInterface) {
					if session, ok := c.Args().At(0).NewSession(); ok {
						log.Printf("New session %+v", session)
					}
				}
			case "pin":
				f = func(c botmeans.ActionContextInterface) {
					msg := c.Args().Raw()[4:]

					if msg != "" {
						if err := DB.Create(&PinnedMsg{Text: msg, SessionId: c.Session().ChatId(), From: c.Session().UserName()}).Error; err != nil {
							log.Println(err)
						}
					}
					c.Finish()
				}
			case "list":
				f = func(c botmeans.ActionContextInterface) {
					msgs := []PinnedMsg{}
					DB.Where("session_id=?", c.Session().ChatId()).Find(&msgs)
					for _, msg := range msgs {
						c.Output().Create("msg", msg)
					}
					c.Finish()
				}
			default:
				s = false
			}

			return
		})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

}

var DBSettings struct {
	DB_ip    string
	DB_port  int16
	DB_login string
	DB_pass  string
	DB_name  string
}

type Config struct {
	DB          *gorm.DB
	Logger      *log.Logger
	BotToken    string
	WebhookHost string
	ListenIP    string
	ListenPort  int16
}

func getConfig(val interface{}) error {
	folderPath, err := osext.ExecutableFolder()

	if err != nil {
		fmt.Println(err)
	}
	jsonBlob, err := ioutil.ReadFile(folderPath + "/settings.json")
	if err != nil {
		fmt.Println(err)
		return err
	}
	if err := json.Unmarshal(jsonBlob, val); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("%+v\n", val)
	return nil
}
