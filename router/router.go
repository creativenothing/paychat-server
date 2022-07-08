package router

import (
	"net/http"

	"github.com/creativenothing/paychat-server/controllers"
	"github.com/gorilla/mux"
)

var Router *mux.Router

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func SetupRouter() {
	Router = mux.NewRouter()
	Router.HandleFunc("/", home)
	Router.HandleFunc("/", controllers.RegisterUser)
}
