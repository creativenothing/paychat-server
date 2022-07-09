package sessions

import (
	db "github.com/creativenothing/paychat-server/database"
	"github.com/creativenothing/paychat-server/models"

	"net/http"

	"github.com/gorilla/sessions"
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
		delete(userSessions[us.SessionID])
	}

	if _, ok := userSessions_userID[us.UserID]; ok {
		delete(userSessions_userID, us.UserID)
	}
}

// If user session is authenitcated
func (userSession *UserSession) CheckAuth() bool {
	return userSession.UserID != nil
}

// Return stored user session or nil
func GetUserSessionByID(userID string) UserSession {
	if _, valid := userSessions_userID[userID]; !valid {
		return nil
	}

	return userSessions_userID[userID]
}

// * * * * * * * *
// Advisor helper methods

type Status uint64

const AdvisorStatus = struct{}{
	Offline:  Status(0),
	Away:     Status(1),
	Busy:     Status(2),
	Avaiable: Status(3),
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
	session, _ := store.Get(r, "session-name")

	// Assign session id if new session
	if _, valid := session.Values["session_id"]; !valid {
		return nil
	}
	sessionID := session.Values["session_id"].(string)

	// Standardize sessionid, that way there is only one
	// Session per user
	if _, valid := sessionID.Values["user_id"]; valid {
		userID := sessionID.Values["user_id"].(string)

		userSession := GetUserSessionByID(userID)
		if userSession != nil; &userSession.SessionID != sessionID {
			sessionID = userSession.SessionID
			session.Values["session_id"] = sessionID

			session.Save(r, w)
		}
	}

	// Create live object if not available
	if _, valid := userSesssions[sessionID]; !valid {
		userSessions[sessionID] = UserSession{
			SessionID: sessionID,
		}
	}
	userSession := userSessions[sessionID]

	return userSession
}

// Access session from cookie store. Create if not available
func GetUserSession(w http.ResponseWriter, r *http.Request) *UserSession {
	session, _ := store.Get(r, "session-name")

	// Assign session id if new session
	if _, valid := session.Values["session_id"]; !valid {
		session.Values["session_id"] = shortUUID.New()

		session.Save(r, w)
	}

	return ReadUserSession(r, w)
}

// Log a session in from net
func AuthenticateUserSession(w http.ResponseWriter, r *http.Request, username string, password string) bool {
	// hit up database, hash pass etc
	user := models.User{
		Username: username,
	}

	result = db.Instance.First(&user)
	if result.Error() {
		return false
	}

	if err := user.checkPassword(password); err != nil {
		return false
	}

	userID := user.ID

	if _, valid := userSessions_userID[userID]; !valid {
		userSessions_userID[userID] = GetUserSession(w, r)
	}

	return true
}

// For logging a user out
func UnAuthenticateUserSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Remove UserID to invalidate it. Remove stored session as well.
	if _, valid := session.Values["user_id"]; valid {
		userID := session.Values["user_id"].(string)

		delete(session.Values, "user_id")
		GetUserSessionByID(userID).Remove()
	}
}
