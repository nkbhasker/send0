package crypto

import (
	"github.com/usesend0/send0/internal/constant"
)

func GenerateOtp() (string, error) {
	return GenerateRandomString(constant.OtpLength, constant.Digits)
}

func ValidateOtp(want, got string) bool {
	if want == "" || got == "" {
		return false
	}

	return want == got
}
