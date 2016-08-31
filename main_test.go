package botmeans

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"os"
	"testing"
)

var DB *gorm.DB
var DBErr error

var bot *MeansBot

func TestMain(t *testing.T) {
	DB, DBErr = gorm.Open("postgres", fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v",
		string(os.Getenv("MEANS_DB_USERNAME")),
		string(os.Getenv("MEANS_DBNAME")),
		""))
	if DBErr != nil {
		t.Fatal(DBErr)
	}
	var err error
	bot, err = New(DB, NetConfig{}, TelegramConfig{BotToken: "262455797:AAHssEUgkKCvvUI88e95OzRZPRc4r8vIlW4",
		WebhookHost: "accounterbot.tuchkov.org",
		SSLCertFile: "./cert.pem"})
	if err != nil {
		t.Fatal(err)
	}
}

func Action1(context *CommandContext) ([]OutputMessage, error) {
	return []OutputMessage{}, nil
}

func Action2(context *CommandContext) ([]OutputMessage, error) {
	return []OutputMessage{}, nil
}

func TestActions(t *testing.T) {
	bot.NewAction("test").Handler(Action1).Handler(Action2)
	t.Logf("%+v", bot.actions["test"])
}
