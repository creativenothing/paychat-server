package database

import (
	"log"

	"github.com/creativenothing/paychat-server/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB
var dbError error

func createTestUser() models.User {

	var u models.User
	u.Username = "admin"
	u.HashPassword("admin")
	return u
}

func Connect(connectionString string) {
	Instance, dbError = gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if dbError != nil {
		log.Fatal(dbError)
		panic("Cannot connect to DB")
	}
	log.Println("Connected to Database!")
}
func Migrate() {
	Instance.AutoMigrate(&models.User{})
	Instance.AutoMigrate(&models.ChatMessage{})

	u := createTestUser()
	Instance.Create(&u)

	log.Println("Database Migration Completed!")
}

func Purge() {
	if Instance.Migrator().HasTable(&models.User{}) {
		Instance.Migrator().DropTable(&models.User{})
	}
}
