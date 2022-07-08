package database

import (
	"log"

	"github.com/creativenothing/paychat-server/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB
var dbError error

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
	//testUser := models.User{Username: "admin", Password: "admin"}
	//Instance.Create(&testUser)
	log.Println("Database Migration Completed!")
}
