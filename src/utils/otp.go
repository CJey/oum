package utils

import (
	"net/url"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func QRLink(text string) string {
	return QRLinkGoogle(text)
}

func QRLinkGoogle(text string) string {
	return "https://www.google.com/chart?chs=400x400&chld=M%7C0&cht=qr&chl=" + text
}

func QRLinkCaoliao(text string) string {
	return "https://cli.im/api/qrcode/code?text=" + text
}

func OTPGenerate(issuer, name string) (secret, text string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: name,
		Period:      60,
		Digits:      otp.DigitsEight,
		Algorithm:   otp.AlgorithmSHA256,
	})
	if err != nil {
		return
	}
	tmp, err := url.QueryUnescape(key.String())
	if err != nil {
		return
	}
	return key.Secret(), url.QueryEscape(tmp), nil
}

func OTPValidate(code, secret string) (bool, error) {
	if len(code) != 8 {
		return false, nil
	}
	return totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    60,
		Skew:      1,
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA256,
	})
}
