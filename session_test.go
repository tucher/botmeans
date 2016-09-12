package botmeans

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"os"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	DB, DBErr := gorm.Open("postgres", fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v",
		string(os.Getenv("MEANS_DB_USERNAME")),
		string(os.Getenv("MEANS_DBNAME")),
		""))
	if DBErr != nil {
		t.Fatal(DBErr)
	}
	DB.AutoMigrate(&Session{})

	s := Session{
		SessionBase: SessionBase{
			TelegramUserID:   123,
			TelegramUserName: "john",
			TelegramChatID:   123,
		},
		UserData:  `{"FF":{"ffuu": 1234}}`,
		CreatedAt: time.Now(),
		db:        DB,
	}

	if err := s.Save(); err != nil {
		t.Errorf("Saving error %v", err)
	}

	loaded, _ := SessionLoader(SessionBase{123, "john", 123, false, false}, DB, 0, nil)
	if loaded.IsNew() != false {
		t.Error("Should be false")
	}
	if loaded.IsLeft() == true {
		t.Error("should be false")
	}
	if loaded.ChatId() != 123 {
		t.Error("should be 123")
	}

	type FF struct {
		Ffuu int
	}
	tdt := FF{}
	loaded.GetData(&tdt)
	if tdt.Ffuu != 1234 {
		t.Error("Should be 1234")
	}

	loaded, _ = SessionLoader(SessionBase{124, "john2", 123, false, false}, DB, 0, nil)
	if loaded.IsNew() != true {
		t.Error("Should be true")
	}
	DB.DropTable(&Session{})
}
