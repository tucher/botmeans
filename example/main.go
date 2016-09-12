package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/tucher/botmeans"
	"log"
	"os"
	// "time"
)

var DB *gorm.DB
var DBErr error

var bot *botmeans.MeansBot
var contextChan chan botmeans.ActionContextInterface

func handlersProvider(id string) (ret botmeans.ActionHandler, ok bool) {
	ok = true
	switch id {
	case "cmd1":
		ret = func(context botmeans.ActionContextInterface) {
			log.Print("Cmd1 Command")
			contextChan <- context
			if len(context.Args()) == 1 {
				if s, _ := context.Args()[0].String(); s == "end" {
					context.Finish()
				}
			}
			context.Output().SimpleText(fmt.Sprintf("Session: (%+v)", context.Session()))
		}

	case "":
		ret = func(context botmeans.ActionContextInterface) {
			log.Println("Empty Command")
			contextChan <- context
			context.Output().SimpleText(fmt.Sprintf("Session: (%+v)", context.Session()))
		}
	default:
		ok = false
	}
	return
}

func main() {
	log.SetFlags(log.Llongfile)
	DB, DBErr = gorm.Open("postgres", fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v",
		string(os.Getenv("MEANS_DB_USERNAME")),
		string(os.Getenv("MEANS_DBNAME")),
		""))
	if DBErr != nil {
		log.Fatal(DBErr)
	}

	DB.DropTable(&botmeans.Session{})
	DB.DropTable(&botmeans.BotMessage{})

	var err error
	bot, err = botmeans.New(DB, botmeans.NetConfig{ListenIP: "0.0.0.0", ListenPort: 7654}, botmeans.TelegramConfig{BotToken: os.Getenv("BOT_TOKEN"),
		WebhookHost: os.Getenv("HOST"),
		SSLCertFile: "../cert.pem"})
	if err != nil {
		log.Fatal(err)
	}

	contextChan = make(chan botmeans.ActionContextInterface)

	bot.Run(handlersProvider, "")
	// enoughChan := time.After(time.Second * 2)
	for {
		select {
		case c := <-contextChan:
			log.Printf("Session: (%+v)", c.Session())
			log.Printf("Args: (%+v)", c.Args())
			// case <-enoughChan:

			// 	return
		}
	}

}
