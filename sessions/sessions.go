package sessions

import (
	"fmt"
	"strconv"

	db "github.com/creativenothing/paychat-server/database"
	"github.com/creativenothing/paychat-server/models"

	"net/http"

	"github.com/gorilla/sessions"
	"github.com/lithammer/shortuuid/v3"
)

const (
	cookieName = "paychat-cookie"
)

// Cookie store to get cookie from user
var store = sessions.NewCookieStore([]byte("super-secret-key"))

// Relates sessionids to live serverside information
var userSessions = map[string]*UserSession{}

// Relates userids to live info
var userSessions_userID = map[string]*UserSession{}

/*
 * SessionID is not null.
 * UserID is null if the user is not logged in
 *
 *
 */
type UserSession struct {
	SessionID string

	UserID string

	Status Status
}

func (us *UserSession) Remove() {
	if _, ok := userSessions[us.SessionID]; ok {
		delete(userSessions, us.SessionID)
	}

	if _, ok := userSessions_userID[us.UserID]; ok {
		delete(userSessions_userID, us.UserID)
	}
}

// For returning user information
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (us *UserSession) UserResponse() UserResponse {
	// Todo: retrieve username from database
	if us.UserID == "" {
		return UserResponse{}
	}

	userIDInt, _ := strconv.Atoi(us.UserID)
	user := models.User{ID: userIDInt}
	db.Instance.Where(&user).First(&user)

	return UserResponse{
		ID:       us.UserID,
		Username: user.Username,
	}

}

// If user session is authenitcated
func (userSession *UserSession) CheckAuth() bool {
	return userSession.UserID != ""
}

// Return stored user session or nil
func GetUserSessionByID(userID string) *UserSession {
	if _, valid := userSessions_userID[userID]; !valid {
		return nil
	}

	return userSessions_userID[userID]
}

// * * * * * * * *
// Advisor helper methods

type Status uint64

var AdvisorStatus = struct {
	Offline   Status
	Away      Status
	Busy      Status
	Available Status
}{
	Offline:   Status(0),
	Away:      Status(1),
	Busy:      Status(2),
	Available: Status(3),
}

func (s Status) String() string {
	switch s {
	case AdvisorStatus.Offline:
		return "offline"
	case AdvisorStatus.Away:
		return "away"
	case AdvisorStatus.Busy:
		return "busy"
	case AdvisorStatus.Available:
		return "available"
	}
	return ""
}

func (us UserSession) AdvisorSetStatus(status Status) {
	us.Status = status

	// I cannot access websocket logic from this file.
	/*us.WidgetHub.broadcast <- []byte(json.Marshall(
	map[string]interface{}{
		"status": status,
	}))*/
}

// * * * * * * * *
// UserSession to http interface

// Access session from cookie store. Return nil if not
func ReadUserSession(w http.ResponseWriter, r *http.Request) *UserSession {
	session, _ := store.Get(r, cookieName)

	// No session if session_id
	if _, valid := session.Values["session_id"]; !valid {
		return nil
	}
	sessionID := session.Values["session_id"].(string)

	// Standardize session by returning usersession
	// Associated with userid when possible
	if _, valid := session.Values["user_id"]; valid {
		userID := session.Values["user_id"].(string)

		userSession := GetUserSessionByID(userID)

		return userSession
	}

	userSession := userSessions[sessionID]

	return userSession
}

// Access session from cookie store. Create if not available
func GetUserSession(w http.ResponseWriter, r *http.Request) *UserSession {
	session, _ := store.Get(r, cookieName)

	// Assign session id if new session
	if _, valid := session.Values["session_id"]; !valid {
		session.Values["session_id"] = shortuuid.New()

		session.Save(r, w)
	}
	sessionID := session.Values["session_id"].(string)

	// This ensures that user sessions are stored in their
	// proper maps
	if userID, valid := session.Values["user_id"].(string); valid {
		// Create at user index if userid exists
		if _, valid := userSessions_userID[userID]; !valid {
			// If there is already a session, use that
			usersession := ReadUserSession(w, r)
			if usersession != nil {
				userSessions_userID[userID] = usersession
			} else {
				// Create new session otherwise
				userSessions_userID[userID] = &UserSession{
					SessionID: sessionID,
					UserID:    userID,
				}
				userSessions[sessionID] = userSessions_userID[userID]
			}
		}
	} else if _, valid := userSessions[sessionID]; !valid {
		// If there is no userid and no session in the structure
		// Create anonymous session
		userSessions[sessionID] = &UserSession{
			SessionID: sessionID,
		}
	}

	return ReadUserSession(w, r)
}

// Log a session in from net
func AuthenticateUserSession(w http.ResponseWriter, r *http.Request, username string, password string) bool {
	session, _ := store.Get(r, cookieName)

	// hit up database
	user := models.User{
		Username: username,
	}

	result := db.Instance.Where(&user).First(&user)
	if result.Error != nil {
		return false
	}

	// Check password
	if err := user.CheckPassword(password); err != nil {
		return false
	}

	userID := fmt.Sprint(user.ID)

	session.Values["user_id"] = userID
	session.Save(r, w)

	return true
}

// For logging a user out
func UnauthenticateUserSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, cookieName)

	// Remove UserID to invalidate it. Remove stored session as well.
	if _, valid := session.Values["user_id"]; valid {
		userID := session.Values["user_id"].(string)

		delete(session.Values, "user_id")
		session.Save(r, w)

		GetUserSessionByID(userID).Remove()
	}
}
