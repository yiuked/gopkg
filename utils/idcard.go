package utils

import (
	"errors"
	"regexp"
	"time"
)

// ParseIdCard 解析身份证返回年龄与生日
func ParseIdCard(idCard string) (age int, birthday string, err error) {
	if len(idCard) <= 18 {
		return 0, "", errors.New("身份证长度错误")
	}

	birthday = idCard[6:14]

	birthdayTime, err := time.Parse("20060102", birthday)
	if err != nil {
		return 0, "", err
	}

	now := time.Now()
	age = now.Year() - birthdayTime.Year()

	if now.YearDay() < birthdayTime.YearDay() {
		age--
	}

	return age, birthday, nil
}

// ValidateIDCard 验证身份证是否合法
func ValidateIDCard(idCard string) bool {
	pattern := `^[1-9]\d{5}(19\d{2}|20\d{2})(0[1-9]|1[012])(0[1-9]|[12]\d|3[01])\d{3}[\dX]$`
	matched, _ := regexp.MatchString(pattern, idCard)
	if !matched {
		return false
	}
	factor := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}
	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(idCard[i]-'0') * factor[i]
	}
	checkCodeIndex := sum % 11
	return string(idCard[17]) == checkCodes[checkCodeIndex]
}
