package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/creativenothing/paychat-server/controllers"
	"github.com/creativenothing/paychat-server/sessions"
	"github.com/creativenothing/paychat-server/websocket"
	"github.com/gorilla/mux"
)

var Router *mux.Router

func preflight(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

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

	fmt.Printf("AUTH %v\n", usersession)

	if !usersession.CheckAuth() {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(usersession.UserResponse())
}

func login(w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("%v\n", usersession)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(usersession.UserResponse())
}

func logout(w http.ResponseWriter, r *http.Request) {
	sessions.UnauthenticateUserSession(w, r)
	usersession := sessions.GetUserSession(w, r)

	fmt.Printf("LOGOUT %v\n", usersession)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(usersession.UserResponse())
}

func token(w http.ResponseWriter, r *http.Request) {
	usersession := sessions.GetUserSession(w, r)
	if !usersession.CheckAuth() {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"type":  "token",
		"token": usersession.GetJWT(),
	})

	//fmt.Println(sessions.GetUserSessionByJWT(usersession.GetJWT()))
}

func wsChatroom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Nil user session to be populated by jwt confirmation
	var usersession *sessions.UserSession

	fmt.Printf("Chat session with %v and %v\n", vars, usersession)

	hub := websocket.GetHub("chat", id)
	var handler websocket.ClientHandler = func(c *websocket.Client, message []byte) {
		fmt.Printf("Chat message with %v and \n%s\n", usersession, string(message))
		messageJSON := map[string]interface{}{}
		if err := json.Unmarshal(message, &messageJSON); err != nil {
			return
		}

		if _, ok := messageJSON["type"].(string); !ok {
			return
		}
		switch messageJSON["type"].(string) {
		case "token":
			if _, ok := messageJSON["token"].(string); !ok {
				return
			}
			token := messageJSON["token"].(string)

			usersession = sessions.GetUserSessionByJWT(token)

		case "chat":
			// Authenticate user
			if !usersession.CheckAuth() {
				return
			}

			// Ensure well formed
			if _, ok := messageJSON["text"].(string); !ok {
				return
			}
			text := messageJSON["text"].(string)

			chat := controllers.NewChatMessage("chat", text, usersession.UserID)

			send, _ := json.Marshal(map[string]interface{}{
				"type":   "chat",
				"id":     chat.ID,
				"userid": usersession.UserID,
				"text":   text,
			})

			c.Broadcast([]byte(send))
			break

		case "typing":
			// Authenticate user
			if !usersession.CheckAuth() {
				return
			}

			// Support dynamic and simple typing indicators.
			// dynamic typing means the incomplete message is shown to the other person while typing.
			if dynamic, ok := messageJSON["dynamic"].(bool); ok {

				// Ensure text is included for dynamic
				if text, ok := messageJSON["text"].(string); dynamic && ok {
					send, _ := json.Marshal(map[string]interface{}{
						"type":    "typing",
						"dynamic": dynamic,
						"text":    text,
					})

					c.MultiCast([]byte(send))

					// Ensure status is included with simple
				} else if status, ok := messageJSON["status"].(bool); ok {
					send, _ := json.Marshal(map[string]interface{}{
						"type":    "typing",
						"dynamic": dynamic,
						"status":  status,
					})

					c.MultiCast([]byte(send))
				}
			}
			break
		}
	}

	websocket.ServeWsWithHandler(hub, w, r, handler)
}

func wsWidget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	advisor := sessions.GetUserSessionByID(id)

	if advisor == nil {
		// Nil session means invalid user
		w.WriteHeader(http.StatusForbidden)
		return
	}

	hub := advisor.AdvisorGetWidgetHub()
	var handler websocket.ClientHandler = func(c *websocket.Client, message []byte) {
		// Do not respond to messages
		return
	}

	client := websocket.ServeWsWithHandler(hub, w, r, handler)
	if client != nil {
		// Update widget status immediately
		messageJSON, _ := json.Marshal(map[string]interface{}{
			"status": advisor.Status.String(),
		})

		client.Send([]byte(messageJSON))
	}
}
func SetupRouter() {
	Router = mux.NewRouter()
	Router.Use(cors)

	Router.HandleFunc("/", preflight).Methods("OPTIONS")

	Router.HandleFunc("/logout", logout).Methods("GET")
	Router.HandleFunc("/login", login).Methods("POST")
	Router.HandleFunc("/signup", controllers.RegisterUser).Methods("POST")
	Router.HandleFunc("/user/{id}", controllers.GetUser).Methods("GET")
	Router.HandleFunc("/user", controllers.GetAllUsers).Methods("GET")
	Router.HandleFunc("/chat/ws", wsChatroom).Methods("GET")
	Router.HandleFunc("/widget/ws", wsWidget).Methods("GET")
	Router.HandleFunc("/auth", auth).Methods("GET")
	Router.HandleFunc("/token", token).Methods("GET")

	Router.HandleFunc("/", home)
}
