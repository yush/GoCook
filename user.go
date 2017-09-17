package main

import (
	"crypto/rand"
	"database/sql"
	"io"
	"log"
	"errors"
	"encoding/base64"
	"html"
)

type SessionManager interface {
	Get(name string) (string, error)
	GetById(id string) string
	Add(name string) error
}

type SessionManagerMem struct {
	Sessions map[string]string
}

func (s *SessionManagerMem) Get(name string) (string, error) {
	val, isPresent := s.Sessions[name]
	if isPresent {
		return val, nil
	} else {
		return "", errors.New("Session not found")
	}
}

func (s *SessionManagerMem) GetById(id string) string {
	unescapedId := html.UnescapeString(id)
	for key, value := range s.Sessions {
		if value == unescapedId {
			return key
		}
	}
	return ""
}

func (s *SessionManagerMem) Add(name string) (string, error) {
	if _, ok := s.Sessions[name]; ok {
		return "", errors.New("Session already exists")
	} else {
		s.Sessions[name] = sessionId()
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
