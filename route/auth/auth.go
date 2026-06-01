package auth

import (
	"errors"
	"net/http"
	"strings"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/session"
	"wholth_go/wholth"
)

type AuthAction int

const (
	ActionRegister AuthAction = iota
	ActionLogin
)

type UnauthorizedPage struct {
	route.HtmlPage
	StatusCode   int
	Username     string
	ErrorMessage string
	Referer      string
}

func DefaultUnauthorizedPage(r *http.Request) UnauthorizedPage {
	page := route.DefaultHtmlPage(r)
	page.Meta.Title = "Аутентикация"
	// meta.Description = "Page does not exist"
	return UnauthorizedPage{
		page,
		http.StatusUnauthorized,
		"",
		"",
		"",
	}
}

func RenderUnauthorizedPage(
	w http.ResponseWriter,
	r *http.Request,
	page UnauthorizedPage,
) {
	route.RenderHtmlTemplatesWithStatus(
		w,
		r,
		page.StatusCode,
		page,
		"templates/index.html",
		"templates/401/content.html",
	)
}

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()

		sess, sess_ok := ctx.Value(session.SessionKey).(session.HttpSession)

		if !sess_ok {
			// wtf how would you even get here
			logger.Info("ADZAM")
			// todo 500 error
			return
		}

		if session.ANON_USERNAME == sess.Username {
			r.Header.Add("Referer", r.URL.RequestURI())
			HandleAuthentication(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HandleAuthentication(w http.ResponseWriter, r *http.Request) {
	page := DefaultUnauthorizedPage(r)
	page.StatusCode = http.StatusOK

	if len(r.Header["Referer"]) > 0 {
		page.Referer = r.Header["Referer"][0]
	}

	defer RenderUnauthorizedPage(w, r, page)
}

func validate_user(r *http.Request) (string, string, error) {
	username := r.PostFormValue("username")

	if "" == username {
		return "", "", errors.New("Заполни имя пользователя")
	}

	trimmed_username := strings.Trim(username, " ")
	if session.ANON_USERNAME == trimmed_username {
		return "", "", errors.New("Недопустимое имя пользователя")
	}

	password := r.PostFormValue("password")

	if "" == password {
		return username, "", errors.New("Заполни пароль")
	}

	return trimmed_username, password, nil
}

// todo pass w as pointer?
func do_auth_RENAME_FUNC(w http.ResponseWriter, r *http.Request, action AuthAction) {
	username, password, user_validation_err := validate_user(r)

	if nil != user_validation_err {
		page := DefaultUnauthorizedPage(r)
		page.Meta.Description = user_validation_err.Error()
		page.StatusCode = http.StatusBadRequest
		page.Username = username
		page.ErrorMessage = user_validation_err.Error()
		RenderUnauthorizedPage(w, r, page)
		return
	}

	var werr error = nil
	var userId string = ""

	if ActionRegister == action {
		userId, werr = wholth.UserRegister(username, password)
	} else {
		userId, werr = wholth.UserAuthenticate(username, password)
	}

	if nil != werr {
		page := DefaultUnauthorizedPage(r)
		page.Meta.Description = werr.Error()
		page.StatusCode = http.StatusOK
		page.Username = username
		page.ErrorMessage = werr.Error()
		RenderUnauthorizedPage(w, r, page)
	} else {
		session.CreateSessionAndSetCookie(w, username, userId)
		referer := r.PostFormValue("referer")
		if len(referer) > 0 {
			http.Redirect(w, r, referer, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func HandleRegistration(w http.ResponseWriter, r *http.Request) {
	do_auth_RENAME_FUNC(w, r, ActionRegister)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	do_auth_RENAME_FUNC(w, r, ActionLogin)
}
