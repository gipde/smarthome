package dao

import (
	"errors"
)

func SaveToken(code string, payload *[]byte) error {
	result := Db.Save(&Token{
		Code:    code,
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
