package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"html"
	"io"
	"log"
	"net/http"
)

type Session struct {
	SessionID string
	UserId    uint
	Email     string
}

type User struct {
	Id       uint
	Username string
	Email    string
	Password string
}

type SessionManager interface {
	Get(name string) (*Session, error)
	GetById(id string) *Session
	Add(ASession Session) error
	CurrentCookieId() (string, error)
	Remove(id string) error
}

type SessionManagerMem struct {
	Sessions map[string]Session
}

func (s *SessionManagerMem) Get(name string) (*Session, error) {
	val, isPresent := s.Sessions[name]
	if isPresent {
		return &val, nil
	} else {
		return nil, errors.New("Session not found")
	}
}

func (s *SessionManagerMem) GetById(id string) (*Session, error) {
	unescapedId := html.UnescapeString(id)
	for _, value := range s.Sessions {
		if value.SessionID == unescapedId {
			return &value, nil
		}
	}
	return nil, errors.New("Session not found")
}

func (s *SessionManagerMem) AddNew(user *User) (*Session, error) {
	_, err := s.Get(user.Email)
	if err == nil {
		return nil, errors.New("Session already exists")
	} else {
		FSession := &Session{sessionId(), user.Id, user.Email}
		s.Sessions[FSession.SessionID] = *FSession
		return FSession, nil
	}
}

func CreateNewUser(db *sql.DB, email string, password string, passConf string) {
	db.Begin()
	defer db.Close()

	if password == passConf {
		var newId int
		db.QueryRow("SELECT MAX(ID) FROM USERS").Scan(&newId)
		_, err := db.Exec("INSERT INTO USERS(ID, USERNAME, EMAIL, PASSWORD) values (?, ?, ?,?)", newId, email, email, password)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetUserByEmail(db *sql.DB, email string) *User {
	db.Begin()
	defer db.Close()

	var user User
	r := db.QueryRow("SELECT ID, USERNAME, EMAIL, PASSWORD FROM USERS WHERE EMAIL = ?", email)
	err := r.Scan(&user.Id, &user.Username, &user.Email, &user.Password)
	if err != nil {
		log.Printf("User %s not found", email)
		log.Print(err)
		return nil
	} else {
		return &user
	}
}

func sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (s *SessionManagerMem) LoggedInUser(req *http.Request) (*Session, error) {
	var sessionid string
	cookie, err := req.Cookie("Cookbook")
	if err != nil {
		return nil, errors.New("No cookies found")
	} else {
		sessionid = html.UnescapeString(cookie.Value)
	}

	CurrentSession, err := s.GetById(sessionid)
	return CurrentSession, err
}

func (s *SessionManagerMem) Remove(id string) {
	delete(s.Sessions, id)
}
