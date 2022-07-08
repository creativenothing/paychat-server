package main

import (
	"fmt"
	"log"
	"net/http"

	db "github.com/creativenothing/paychat-server/database"
	r "github.com/creativenothing/paychat-server/router"
	ws "github.com/creativenothing/paychat-server/websocket"
)

var addr = "http://localhost:8080"

func main() {

	db.Connect("test.db")
	db.Migrate()

	r.SetupRouter()

	ws.Instance.Run()

	err := http.ListenAndServe(addr, r.Router)

	if err != nil {
		log.Fatal("HTTP Server Error: ", err)
	}
	fmt.Printf("Server running on port %s!\n", addr)
}
