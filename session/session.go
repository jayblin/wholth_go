package session

import (
	"context"
	"strconv"

	// "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	// "fmt"
	// "math/big"
	"net/http"
	// "strconv"
	"strings"
	"time"
	"wholth_go/logger"
	"wholth_go/secret"
	"wholth_go/wholth"
)

type HttpSession struct {
	Expires         time.Time
	ExpiresStr      string
	Username        string
	UserId          string
	Value           string
	IsAuthenticated bool
}

var G_sessions = make(map[string]HttpSession)

const ANON_USERNAME = "anon"
const SESSION_COOKIE_NAME = "sesid"

func persist_session_from_cookie(sesid *http.Cookie) (HttpSession, bool) {
	var result = HttpSession{}

	parts := strings.SplitN(sesid.Value, ":", 3)

	if 3 != len(parts) || ANON_USERNAME == parts[0] || 0 == len(parts[0]) {
		return result, false
	}

	userId, userExistsErr := wholth.UserExists(parts[0])

	if nil != userExistsErr {
		return result, false
	}

	expiresAt, expiresAtErr := strconv.ParseInt(parts[1], 10, 64)

	if nil != expiresAtErr {
		return result, false
	}

	expiresAtTime := time.UnixMicro(expiresAt)
	expected := generate_session_value(parts[0], userId, expiresAtTime)

	if expected != sesid.Value {
		return result, false
	}

	return create_session(sesid.Value, expiresAtTime, parts[0], userId), true
}

func get_session(r *http.Request) (HttpSession, error) {
	now := time.Now().Unix()
	sesid, err := r.Cookie(SESSION_COOKIE_NAME)
	sess := HttpSession{}

	if nil != err || "" == sesid.Value {
		return sess, errors.New("ABZUG")
	}

	var elem, ok = G_sessions[sesid.Value]

	if !ok {
		// if user has good looking session cookie - then
		// persist it in the server.
		// todo add a blacklist for compormised cookies?
		elem, ok = persist_session_from_cookie(sesid)

		if !ok {
			return sess, errors.New("ABSCHIED")
		}
	}

	if elem.Expires.Unix() <= now {
		delete(G_sessions, sesid.Value)
		return sess, errors.New("AVANIA")
	}

	return elem, nil
}

func generate_session_value(userName string, userId string, expiresAt time.Time) string {
	sessSecret := secret.GetSessionSecret()
	if len(sessSecret) < 32 {
		panic("AKANTHOS")
	}

	exp := strconv.FormatInt(expiresAt.UnixMicro(), 10)
	sum := sha256.Sum256([]byte(
		userName + userId + exp + sessSecret))
	return userName + ":" + exp + ":" + hex.EncodeToString(sum[:])
}

func create_session(
	value string,
	expiresAt time.Time,
	userName string,
	userId string,
) HttpSession {
	sess := HttpSession{
		Expires:         expiresAt,
		ExpiresStr:      expiresAt.Format(time.DateTime),
		Username:        userName,
		UserId:          userId,
		Value:           value,
		IsAuthenticated: "" != userName && userName != ANON_USERNAME,
	}

	G_sessions[value] = sess

	return sess
}

func CreateSessionAndSetCookie(w http.ResponseWriter, username string, userId string) (HttpSession, error) {
	domain := secret.GetDomain()

	if "" == domain {
		panic("EKAMISTOS")
	}

	inAMonth := time.Now().Add(780 * time.Hour)

	sess_value := generate_session_value(username, userId, inAMonth)

	sess := create_session(sess_value, inAMonth, username, userId)

	// https://www.golinuxcloud.com/http-set-cookie-golang/
	sess_cookie := http.Cookie{}
	sess_cookie.Name = SESSION_COOKIE_NAME
	sess_cookie.Value = sess_value
	sess_cookie.Path = "/"
	// todo get from env?
	// sess_cookie.Domain = "localhost" //potential vulnerability?
	sess_cookie.Domain = domain
	sess_cookie.Expires = inAMonth
	// sess_cookie.MaxAge = 780 * 60 * 60
	// sess_cookie.Secure = true
	sess_cookie.HttpOnly = true
	// sess_cookie.SameSite = ???

	http.SetCookie(w, &sess_cookie)

	return sess, nil
}

func GenerateScrfToken(session HttpSession) string {
	csrfSecret := secret.GetCsrfSecret()
	if len(csrfSecret) < 32 {
		logger.Alert("AGRAFE")
		return ""
	}

	sum := sha256.Sum256([]byte(csrfSecret + session.Value))

	return hex.EncodeToString(sum[:])
}

var SessionKey *HttpSession

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sess, err = get_session(r)

		if nil != err {
			sess, err = CreateSessionAndSetCookie(w, ANON_USERNAME, "")
		}

		if nil != err {
			// todo serve 500 page
			return
		}

		ctx := context.WithValue(r.Context(), SessionKey, sess)
		// r.Context().Value()
		// http.NewRequestWithContext()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type csrfTokenKeyType struct{}

var CsrfTokenKey *csrfTokenKeyType

func CsrfTokenGeneratorMiddleware(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()

		sess, sess_ok := ctx.Value(SessionKey).(HttpSession)

		if sess_ok {
			token := GenerateScrfToken(sess)
			ctx = context.WithValue(ctx, CsrfTokenKey, token)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
	return SessionMiddleware(handler)
}

func CsrfTokenValidatorMiddleware(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()

		sess, sess_ok := ctx.Value(SessionKey).(HttpSession)

		if !sess_ok {
			logger.Info("ADAMANT")
			// todo 400 error - bad session
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Нет сессии!"))
			return
		}

		token := r.PostFormValue("csrf-token")

		if "" == token {
			logger.Info("ABULIA")
			// todo 400 error - bad csrf token
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Нет CSRF-токена!"))
			return
		}

		expected_token := GenerateScrfToken(sess)

		if expected_token != token {
			logger.Info("AUSTERIA")
			// todo 400 error - bad csrf token
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("CSRF-токен не соотв. ожиадаемому!"))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})

	return SessionMiddleware(handler)
}

func GetCsrfToken(r *http.Request) string {
	val, ok := r.Context().Value(CsrfTokenKey).(string)

	if !ok {
		return ""
	}

	return val
}

func GetSession(r *http.Request) HttpSession {
	return r.Context().Value(SessionKey).(HttpSession)
}
