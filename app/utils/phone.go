package utils

import "github.com/nyaruka/phonenumbers"

func ValidatePhoneNum(phoneNum string) bool {
	parsedNum, err := phonenumbers.Parse(phoneNum, "ZZ")
	if err != nil {
		return false
	}

	return phonenumbers.IsValidNumber(parsedNum)
}
