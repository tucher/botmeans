package botmeans

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"os"
	"testing"
	"time"
)

func TestBotMessage(t *testing.T) {
	DB, DBErr := gorm.Open("postgres", fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v",
		string(os.Getenv("MEANS_DB_USERNAME")),
		string(os.Getenv("MEANS_DBNAME")),
		""))
	if DBErr != nil {
		t.Fatal(DBErr)
	}
	DB.AutoMigrate(&BotMessage{})

	msg := BotMessage{
		TelegramMsgID:  123,
		TelegramChatID: 123,
		UserData:       `{"FF":{"ffuu": 1234}}`,
		Timestamp:      time.Now(),
		db:             DB,
	}

	if err := msg.Save(); err != nil {
		t.Errorf("Saving error %v", err)
	}

	loaded := BotMessageDBLoader(123, 123, "2", DB)
	if loaded.CallbackID() != "2" {
		t.Error("Should be '2'")
	}
	type FF struct {
		Ffuu int
	}
	tdt := FF{}
	loaded.GetData(&tdt)
	if tdt.Ffuu != 1234 {
		t.Error("Should be 1234")
	}
	DB.DropTable(&BotMessage{})

}
