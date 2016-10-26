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
			if context.Args().Count() == 1 {
				if s, _ := context.Args().At(0).String(); s == "end" {
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
	log.SetFlags(log.Lshortfile)
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
			log.Printf("Args:")
			for i := 0; i < c.Args().Count(); i++ {
				arg := c.Args().At(i)
				if s, ok := arg.Mention(); ok {
					log.Println("   ", i, ": Mention ", s)
					// log.Println(s.Save())
				}

				if s, ok := arg.NewSession(); ok {
					log.Println("   ", i, ": NewSession: ", s)
				}
				if s, ok := arg.LeftSession(); ok {
					log.Println("   ", i, ": LeftSession: ", s)
				}
				if s, ok := arg.ComeSession(); ok {
					log.Println("   ", i, ": ComeSession: ", s)

				}

				log.Println("   ", i, ": Arg ", arg)

			}

			log.Printf("\n")
			// case <-enoughChan:

			// 	return
		}
	}

}
