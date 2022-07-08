package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/creativenothing/paychat-server/database"
	"github.com/creativenothing/paychat-server/models"
)

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
	record := database.Instance.Create(&u)
	if record.Error != nil {
		http.Error(w, record.Error.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	resp := map[string]interface{}{"username": u.Username, "id": u.ID}
	json.NewEncoder(w).Encode(resp)

}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	database.Instance.Find(&users)

	resp, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, string(resp))

}

// PUT user
// DELETE user
