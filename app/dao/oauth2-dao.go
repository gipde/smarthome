package dao

import (
	"errors"
	"github.com/ory/fosite"
	"github.com/revel/revel"
	"time"
)

func SaveToken(signature, tokenid string, tokentype fosite.TokenType, expiry time.Time, payload *[]byte) error {
	result := Db.Save(&Token{
		Signature: signature,
		TokenID:   tokenid,
		TokenType: tokentype,
		Expiry:    expiry,
		PayLoad:   *payload,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("something went wrong, affected rows != 1")
	}
	return nil
}

func GetTokenBySignature(signature string) *[]byte {
	var aToken Token
	result := Db.First(&aToken, "signature = ?", signature)
	if result.RowsAffected == 0 {
		return nil
	}
	return &aToken.PayLoad
}

func GetTokenByTokenID(tokenid string, tokentype fosite.TokenType) *[]byte {
	var aToken Token
	result := Db.First(&aToken, "token_id = ?  and token_type=?", tokenid, tokentype)
	if result.RowsAffected == 0 {
		return nil
	}
	return &aToken.PayLoad
}

func DeleteToken(code string) error {
	var token Token
	result := Db.Delete(&token, "signature = ?", code)
	if result.RowsAffected == 0 {
		return errors.New("no token found")
	}
	return nil
}

func DeleteTokenByTokenID(tokenid string, tokentype fosite.TokenType) error {
	var aToken Token
	result := Db.Delete(&aToken, "token_id = ?  and token_type=?", tokenid, tokentype)
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
