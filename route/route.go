package route

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
		Lang:        "en-US",
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
}

func DefaultHtmlPage(r *http.Request) HtmlPage {
	return HtmlPage{
		DefaultPageMeta(r),
		DefaultSessionMeta(r),
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
		return nil, errors.New("ABAZI " + err.Error())
	}

	if secret.GetUseTemplateCache() {
		G_templateMap[hash] = tmpl
	}

	return tmpl, nil
}

func RenderHtmlTemplates(
	w http.ResponseWriter,
	r *http.Request,
	data any,
	filenames ...string,
) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := parse_templates(filenames...)

	if nil != err {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)

	if nil != err {
		logger.Error("AGHA " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
