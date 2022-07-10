package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/creativenothing/paychat-server/database"
	"github.com/creativenothing/paychat-server/models"
	"github.com/gorilla/mux"
)

func readJSON(r *http.Request) {

}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var u models.User

	body := json.NewDecoder(r.Body)

	if err := body.Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := u.HashPassword(u.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result := db.Instance.Create(&u)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	resp := map[string]interface{}{"username": u.Username, "id": u.ID}
	json.NewEncoder(w).Encode(resp)

}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	u := models.User{ID: id}

	result := db.Instance.First(&u)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	fmt.Printf("%+v\n", u)
	json.NewEncoder(w).Encode(&u)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	db.Instance.Find(&users)

	resp, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, string(resp))

}

// PUT user
// DELETE user
