package route

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	// "errors"
	"html/template"
	"net/http"
	"wholth_go/logger"
	"wholth_go/secret"
	"wholth_go/session"
)

type PageMeta struct {
	Title       string
	Description string
	Lang        string
}

func DefaultPageMeta(r *http.Request) PageMeta {
	return PageMeta{
		Title:       "",
		Description: "",
		Lang:        "ru-RU",
	}
}

type SessionMeta struct {
	session.HttpSession
	CsrfToken string
}

func DefaultSessionMeta(r *http.Request) SessionMeta {
	return SessionMeta{
		session.GetSession(r),
		session.GetCsrfToken(r),
	}
}

type HtmlPage struct {
	Meta    PageMeta
	Session SessionMeta
	Version string
}

func DefaultHtmlPage(r *http.Request) HtmlPage {
	return HtmlPage{
		Meta:    DefaultPageMeta(r),
		Session: DefaultSessionMeta(r),
		Version: secret.GetVersion(),
	}
}

var G_templateMap = make(map[string]*template.Template)

func parse_templates(filenames ...string) (*template.Template, error) {
	var hash string

	if secret.GetUseTemplateCache() {
		var fns = ""
		for _, filename := range filenames {
			fns += filename
		}
		sum := sha256.Sum256([]byte(fns))
		hash = hex.EncodeToString(sum[:])

		var tmpl, found = G_templateMap[hash]

		if found {
			return tmpl, nil
		}
	}

	tmpl, err := template.ParseFiles(filenames...)

	if nil != err {
		return nil, err
	}

	if secret.GetUseTemplateCache() {
		G_templateMap[hash] = tmpl
	}

	return tmpl, nil
}

func RenderHtmlTemplatesWithStatus(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	data any,
	filenames ...string,
) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := parse_templates(filenames...)

	if nil != err {
		logger.Error("ABAZI " + err.Error())
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	err = tmpl.Execute(w, data)

	if nil != err {
		logger.Error("AGHA " + err.Error())
		logger.Error("AGHA " + strings.Join(filenames, ","))
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func RenderHtmlTemplates(
	w http.ResponseWriter,
	r *http.Request,
	data any,
	filenames ...string,
) {
	RenderHtmlTemplatesWithStatus(w, r, 200, data, filenames...)
}
