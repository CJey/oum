package api

import (
	"fmt"
)

const (
	API_UNDEFINED = "undefined"

	ERR_UNEXPECTED   = 100
	ERR_ARG_NOTFOUND = 101
	ERR_ARG_EMPTY    = 102
	ERR_ARG_INVALID  = 103

	ERR_API_UNDEFINED = 200
	ERR_API_INVALID   = 201
	ERR_OUTPUT        = 202
)

func getSMSSign(sign string) string {
	switch sign {
	case "来客":
		return "【来客】"
	case "来客之家":
		return "【来客之家】"
	case "来客pos机":
		return "【来客pos机】"
	}
	return ""
}

func register_message(tp, str, sign string) string {
	switch tp {
	case "sms":
		return fmt.Sprintf(`您当前正在进行注册操作,验证码为: %s (30分钟内有效，请您在页面中输入以完成验证，验证码请勿泄露。如非本人操作请忽略。)如有其它问题，请拨打来客官方客服电话4008732188 %s`, str, sign)
	case "email":
		//TODO
	case "call":
		//TODO
	}
	return ""
}

func bind_message(tp, str, sign string) string {
	switch tp {
	case "sms":
		return fmt.Sprintf(`您当前正在进行绑定操作,验证码为: %s (30分钟内有效，请您在页面中输入以完成验证，验证码请勿泄露。如非本人操作请忽略。)如有其它问题，请拨打来客官方客服电话4008732188 %s`, str, sign)
	case "email":
		//TODO
	case "call":
		//TODO
	}
	return ""
}

func unbind_message(tp, str, sign string) string {
	switch tp {
	case "sms":
		return fmt.Sprintf(`您当前正在进行解除绑定操作,验证码为: %s (30分钟内有效，请您在页面中输入以完成验证，验证码请勿泄露。如非本人操作请忽略。)如有其它问题，请拨打来客官方客服电话4008732188 %s`, str, sign)
	case "email":
		//TODO
	case "call":
		//TODO
	}
	return ""
}

func reset_message(tp, str, sign string) string {
	switch tp {
	case "sms":
		return fmt.Sprintf(`您当前正在进行重置密码操作,验证码为: %s (30分钟内有效，请您在页面中输入以完成验证，验证码请勿泄露。如非本人操作请忽略。)如有其它问题，请拨打来客官方客服电话4008732188 %s`, str, sign)
	case "email":
		//TODO
	case "call":
		//TODO
	}
	return ""
}

func safe_message(tp, str, sign string) string {
	switch tp {
	case "sms":
		return fmt.Sprintf(`您当前正在安全登录操作,验证码为: %s (30分钟内有效，请您在页面中输入以完成验证，验证码请勿泄露。如非本人操作请忽略。)如有其它问题，请拨打来客官方客服电话4008732188 %s`, str, sign)
	case "email":
		//TODO
	case "call":
		//TODO
	}
	return ""
}
