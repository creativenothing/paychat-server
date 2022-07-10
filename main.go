package main

import (
	"log"
	"net/http"

	db "github.com/creativenothing/paychat-server/database"
	r "github.com/creativenothing/paychat-server/router"
	ws "github.com/creativenothing/paychat-server/websocket"
)

var addr = "localhost:8080"

func main() {

	db.Connect("test.db")
	db.Purge()
	db.Migrate()

	r.SetupRouter()

	go ws.Instance.Run()

	err := http.ListenAndServe(addr, r.Router)
	if err != nil {
		log.Fatal("HTTP Server Error: ", err)
	}
	log.Printf("Server running on port %s!\n", addr)
}
