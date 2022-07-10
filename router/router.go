package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/creativenothing/paychat-server/controllers"
	"github.com/creativenothing/paychat-server/sessions"
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

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all users here"))
}

func auth(w http.ResponseWriter, r *http.Request) {
	usersession := sessions.GetUserSession(w, r)

	if !usersession.CheckAuth() {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(usersession.UserResponse())
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Retrieve parameters from http body
	message := map[string]interface{}{}
	json.NewDecoder(r.Body).Decode(&message)
	username, userok := message["username"].(string)
	password, passok := message["password"].(string)

	fmt.Printf("LOGIN:\n%v\n", message)

	// Malformed request if fields do not exist
	if !passok || !userok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to authenticate and report failure
	authenticated := sessions.AuthenticateUserSession(w, r, username, password)

	if !authenticated {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	usersession := sessions.GetUserSession(w, r)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(usersession.UserResponse())
}

func logout(w http.ResponseWriter, r *http.Request) {
	sessions.UnauthenticateUserSession(w, r)
	usersession := sessions.GetUserSession(w, r)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(usersession.UserResponse())
}

func SetupRouter() {
	Router = mux.NewRouter()
	Router.Use(cors)
	Router.HandleFunc("/", home)
	Router.HandleFunc("/logout", logout).Methods("GET")
	Router.HandleFunc("/login", login).Methods("POST", "OPTIONS")
	Router.HandleFunc("/user", controllers.RegisterUser).Methods("POST")
	Router.HandleFunc("/user/{id}", controllers.GetUser).Methods("GET")
	Router.HandleFunc("/user", controllers.GetAllUsers).Methods("GET")
	Router.HandleFunc("/auth", auth).Methods("GET")

}
