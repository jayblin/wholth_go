package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"

import (
	"errors"

	// "fmt"

	"wholth_go/secret"
)

func UserRegister(username string, password string) (string, error) {
	if !secret.GetAllowRegistration() {
		return "", errors.New("Регистрация в данный момент не разрешена!")
	}

	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	wuser := C.wholth_entity_user_init()
	wuser.name = toStrView(username)
	wuser.locale_id = toStrView(DEFAULT_LOCALE_ID)
	wpassword := toStrView(password)
	werr := C.wholth_em_user_insert(&wuser, wpassword, scratch)

	if !C.wholth_error_ok(&werr) {
		return "", errors.New(toStr(werr.message))
	}

	return toStr(wuser.id), nil
}

func UserExists(username string) (string, error) {
	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	wusername := toStrView(username)
	var wid = C.wholth_StringView{}

	werr := C.wholth_em_user_exists(wusername, &wid, scratch)

	if !C.wholth_error_ok(&werr) {
		return "", errors.New(toStr(werr.message))
	}

	return toStr(wid), nil
}

func UserAuthenticate(username string, password string) (string, error) {
	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	wuser := C.wholth_entity_user_init()
	wuser.name = toStrView(username)
	wpassword := toStrView(password)
	werr := C.wholth_em_user_authenticate(&wuser, wpassword, scratch)

	if !C.wholth_error_ok(&werr) {
		return "", errors.New(toStr(werr.message))
	}

	return toStr(wuser.id), nil
}
