package dao

import (
	"errors"
	"github.com/revel/revel"
	"time"
)

func SaveToken(code string, expiry time.Time, payload *[]byte) error {
	result := Db.Save(&Token{
		Code:    code,
		Expiry:  expiry,
		PayLoad: *payload,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("something went wrong, affected rows != 1")
	}
	return nil
}

func GetToken(code string) *[]byte {
	var aToken Token
	result := Db.First(&aToken, "code = ?", code)
	if result.RowsAffected == 0 {
		return nil
	}
	return &aToken.PayLoad
}

func DeleteToken(code string) error {
	var token Token
	result := Db.Delete(&token, "code = ?", code)
	if result.RowsAffected == 0 {
		return errors.New("no token found")
	}
	return nil
}

func GetAllTokens() *[]Token {
	var tokens []Token
	Db.Find(&tokens)
	return &tokens
}

func CleanExpiredTokens() {
	var tokens Token
	result := Db.Delete(&tokens, "Expiry < ?", time.Now().UTC())
	if result.RowsAffected > 0 {
		revel.AppLog.Infof("Deleted %d expired Tokens", result.RowsAffected)
	}
}
