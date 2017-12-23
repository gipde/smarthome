package dao

import (
	"github.com/jinzhu/gorm"
	"github.com/ory/fosite"
	"time"
)

type User struct {
	gorm.Model
	UserID         string
	Name           string
	Password       []byte
	DevicePassword []byte
	Devices        []Device
	Authorizations []AuthorizeEntry
}

type Device struct {
	gorm.Model
	UserID      uint
	Name        string
	Description string
	Producer    string
	DeviceType  int
	State       string // eg a json fragment
	Connected   bool
}

type AuthorizeEntry struct {
	gorm.Model
	UserID       uint
	AppID        string
	RefreshToken string
}

type Token struct {
	gorm.Model
	Expiry    time.Time
	TokenID   string
	TokenType fosite.TokenType
	Signature string
	PayLoad   []byte
}
