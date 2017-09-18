package main

import (
	"crypto/rand"
	"database/sql"
	"io"
	"log"
	"errors"
	"encoding/base64"
	"html"
	"net/http"
)

type Session struct {
	SessionID string
	UserId    int
	UserName  string
}

type SessionManager interface {
	Get(name string) (Session, error)
	GetById(id string) Session
	Add(ASession Session) error
	CurrentCookieId() (string, error)
}

type SessionManagerMem struct {
	Sessions map[string]Session
}

func (s *SessionManagerMem) Get(name string) (Session, error) {
	val, isPresent := s.Sessions[name]
	if isPresent {
		return val, nil
	} else {
		return Session{}, errors.New("Session not found")
	}
}

func (s *SessionManagerMem) GetById(id string) Session {
	unescapedId := html.UnescapeString(id)
	for _, value := range s.Sessions {
		if value.SessionID == unescapedId {
			return value
		}
	}
	return Session{}
}

func (s *SessionManagerMem) Add(name string) (Session, error) {
	if _, ok := s.Sessions[name]; ok {
		return Session{}, errors.New("Session already exists")
	} else {
		s.Sessions[name] = Session{sessionId(), 0, name}
		return s.Sessions[name], nil
	}
}

func CreateNewUser(db *sql.DB, email string, password string, passConf string) {
	db.Begin()
	defer db.Close()

	if password == passConf {
		_, err := db.Exec("INSERT INTO USERS(EMAIL, PASSWORD) values (?,?)", email, password)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (s *SessionManagerMem) LoggedInUser(req *http.Request) (Session, error) {
	var sessionid string
	cookie, err := req.Cookie("Cookbook")
	if err != nil {
		return Session{}, errors.New("No cookies found")
	} else {
		sessionid = html.UnescapeString(cookie.Value)
	}

	CurrentSession := s.GetById(sessionid)
	if CurrentSession.UserName == "" {
		return Session{}, errors.New("No logged in user")
	} else {
		return CurrentSession, nil
	}
}
