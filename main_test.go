package botmeans

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"os"
	"testing"
	"time"
)

var DB *gorm.DB
var DBErr error

var bot *MeansBot2

func TestMain(t *testing.T) {
	DB, DBErr = gorm.Open("postgres", fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v",
		string(os.Getenv("MEANS_DB_USERNAME")),
		string(os.Getenv("MEANS_DBNAME")),
		""))
	if DBErr != nil {
		t.Fatal(DBErr)
	}
	var err error
	bot, err = New2(DB, NetConfig{ListenIP: "0.0.0.0", ListenPort: 7654}, TelegramConfig{BotToken: os.Getenv("BOT_TOKEN"),
		WebhookHost: os.Getenv("HOST"),
		SSLCertFile: "./cert.pem"})
	if err != nil {
		t.Fatal(err)
	}

	contextChan := make(chan ActionContextInterface)
	handlersProvider := func(id string) (ret ActionHandler, ok bool) {
		ok = true
		switch id {
		case "cmd1":
			ret = func(context ActionContextInterface) {
				t.Log("Cmd1 Command")
				contextChan <- context
			}

		case "":
			ret = func(context ActionContextInterface) {
				t.Log("Empty Command")
				contextChan <- context
			}
		default:
			ok = false
		}
		return
	}
	bot.Run(handlersProvider, "")
	enoughChan := time.After(time.Second * 2)
	for {
		select {
		case c := <-contextChan:
			t.Logf("Session: (%+v)", c.Session())
			t.Logf("Args: (%+v)", c.Args())
		case <-enoughChan:
			DB.DropTable(&Session{})
			DB.DropTable(&BotMessage{})
			return
		}
	}

}
