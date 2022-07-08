package router

import (
	"encoding/json"
	"net/http"

	"github.com/creativenothing/paychat-server/controllers"
	"github.com/gorilla/mux"
)

var Router *mux.Router

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token")
		//w.Header().Set("Access-Control-Allow-Headers", "true")
		h.ServeHTTP(w, r)

	})
}

var user = map[string]interface{}{"id": 1}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all users here"))
}

func auth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(user)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	user["id"] = 1
	json.NewEncoder(w).Encode(user)

}
func logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	user["id"] = nil
	json.NewEncoder(w).Encode(user)

}

func SetupRouter() {
	Router = mux.NewRouter()
	Router.Use(cors)
	Router.HandleFunc("/", home)
	Router.HandleFunc("/logout", logout).Methods("GET")
	Router.HandleFunc("/login", login).Methods("POST", "OPTIONS")
	Router.HandleFunc("/user", controllers.RegisterUser).Methods("POST")
	Router.HandleFunc("/user/{id}", controllers.GetUser).Methods("GET")
	Router.HandleFunc("/user", getUsers).Methods("GET")
	Router.HandleFunc("/auth", auth).Methods("GET")

}
