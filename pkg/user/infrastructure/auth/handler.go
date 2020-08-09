package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ilya-shikhaleev/arch-course/pkg/user/app/user"
)

type SessionService struct {
	sessions map[string]session
	repo     user.Repository
	encoder  user.PassEncoder
}

func NewSessionService(repo user.Repository, encoder user.PassEncoder) *SessionService {
	sessions := make(map[string]session)
	return &SessionService{sessions: sessions, repo: repo, encoder: encoder}
}

type session struct {
	id        string
	login     string
	email     string
	firstName string
	lastName  string
}

const sessionCookie = "sid"

func (service *SessionService) AuthHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie(sessionCookie)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, ok := service.sessions[sessionID.Value]
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("X-User-Id", session.id)
	w.Header().Set("X-Login", session.login)
	w.Header().Set("X-Email", session.email)
	w.Header().Set("X-First-Name", session.firstName)
	w.Header().Set("X-Last-Name", session.lastName)
	w.WriteHeader(http.StatusOK)
}

func (service *SessionService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var data struct {
		Login    string
		Password string
	}
	err := decoder.Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, err.Error())
		return
	}

	u, err := service.repo.FindByUsername(data.Login)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, err.Error())
		return
	}

	if service.encoder.Encode(data.Password) != u.EncodedPass {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Invalid password")
		return
	}

	sessionID := generateSessionID()
	service.sessions[sessionID] = session{
		id:        string(u.ID),
		login:     u.Username,
		email:     string(u.Email),
		firstName: u.FirstName,
		lastName:  u.LastName,
	}

	c := &http.Cookie{
		Name:    sessionCookie,
		Value:   sessionID,
		Path:    "/",
		Expires: time.Now().Local().Add(time.Minute * 15),

		HttpOnly: true,
	}
	http.SetCookie(w, c)
	w.WriteHeader(http.StatusOK)
}

func (service *SessionService) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	c := &http.Cookie{
		Name:    sessionCookie,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	}
	http.SetCookie(w, c)
	w.WriteHeader(http.StatusOK)

	if sessionID, err := r.Cookie(sessionCookie); err == nil {
		delete(service.sessions, sessionID.Value)
	}
}

func generateSessionID() string {
	return uuid.New().String()
}
