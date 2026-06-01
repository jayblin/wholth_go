package exercise

import (
	"fmt"
	"net/http"
	"time"

	// "slices"

	"wholth_go/container"
	"wholth_go/logger"
	"wholth_go/session"

	// "wholth_go/logger"
	"wholth_go/route"
	"wholth_go/util"
	"wholth_go/wholth"
)

type MapElement struct {
	// NutrientAmountSum float64
	Values []wholth.ExerciseLog
}

type ListExerciseLogPage struct {
	route.HtmlPage
	util.EntityAliasAware[wholth.ExerciseLog]
	util.PaginatableList[wholth.ExerciseLog]
	util.Status
	Q                 string
	Groups            []string
	Map               map[string]MapElement
	From              string
	To                string
	Types             []wholth.ExerciseType
	PostForm          PostExerciseLogForm
	DummyExerciseList struct {
		util.PaginatableList[wholth.Exercise]
	}
}

func ListExerciseLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// TODO adapt to user's timezone
	var from, fromErr = time.Parse(wholth.DateFormat(), q.Get("from"))
	if nil != fromErr {
		from = time.Now().Add((time.Hour * 5) - (time.Hour * 48))
	}

	var to, toErr = time.Parse(wholth.DateFormat(), q.Get("to"))
	if nil != toErr {
		to = time.Now().Add(time.Hour * 5)
	}

	list, err, sev := wholth.FetchExerciseLogList(r, session.GetSession(r).UserId, from, to)

	if nil != err {
		container.Log(r, sev, "[ListExerciseLogs]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	mapped := make(map[string]MapElement)
	groups := make([]string, 0)

	size := len(list.Values)

	if size > 0 {
		for i := size; i > 0; {
			i--
			value := list.Values[i]

			// 1234567890123456789
			// 2025-12-14T23:48:53
			var grp = ""

			if len(value.CreatedAt) >= 10 {
				grp = value.CreatedAt[0:10]
			}

			// value.CreatedAt = value.CreatedAt[11:]

			entry, ok := mapped[grp]

			if !ok {
				entry = MapElement{
					Values: make([]wholth.ExerciseLog, 0),
				}
				mapped[grp] = entry
				// groups = append(groups, grp)
				groups = append([]string{grp}, groups...)
			}

			entry.Values = append(entry.Values, value)

			mapped[grp] = entry
		}
	}

	types, err, sev := wholth.FetchExerciseTypeList(r)

	if nil != err {
		container.Log(r, sev, "[ListExerciseLogs]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		// return
	}

	page := ListExerciseLogPage{}
	page.PaginatableList = list
	page.Types = types
	page.Groups = groups
	page.Map = mapped
	page.From = from.Format(wholth.DateFormat())
	page.To = to.Format(wholth.DateFormat())
	page.HtmlPage = route.DefaultHtmlPage(r)
	// page.PostForm = wholth.PostFoodsFormDefault()
	page.Meta.Title = fmt.Sprintf(
		"Логи упражнений // Страница %d",
		page.PaginatableList.Pagination.PageCurrent,
	)
	page.Meta.Description = ""
	page.PostForm = PostExerciseLogForm{
		Id: "",
		// TODO adapt to user's timezone
		CreatedAt: time.Now().Add(time.Hour * 5).Format(wholth.DateFormat()),
	}

	as_subdoc := q.Get("as_subdoc")

	if "" != as_subdoc {
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/exercise_log/get/as_subdoc.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",

			"templates/exercise_log/get/form.html",
		)
	} else {
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/index.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",
			"templates/utils/toggleable.html",

			"templates/exercise_log/post/form.html",
			"templates/exercise_log/post/result.html",

			"templates/exercise_log/get/content.html",
			"templates/exercise_log/get/form.html",
		)
	}
}

type PostExerciseLogForm struct {
	Id        string
	CreatedAt string
	TypeId    string
	Value     string
	Exercise  wholth.Exercise
	// BodyParts util.PaginatableList[wholth.BodyPart]
}

type PostExerciseLogPage struct {
	route.HtmlPage
	util.Status
	PostForm PostExerciseLogForm
}

func parseExerciseLogForm(r *http.Request) PostExerciseLogForm {
	r.ParseForm()

	createdAt, createdAtErr := time.Parse(wholth.DateFormat(), r.PostForm.Get("created_at"))

	form := PostExerciseLogForm{
		Id:        r.PostForm.Get("id"),
		CreatedAt: "",
		TypeId:    r.PostForm.Get("type_id"),
		Value:     r.PostForm.Get("value"),
		Exercise: wholth.Exercise{
			Id: r.PostForm.Get("exercise_id"),
		},
	}

	if nil == createdAtErr {
		form.CreatedAt = createdAt.Format(wholth.DateFormat())
	}

	return form
}

func saveExerciseLog(r *http.Request, form *PostExerciseLogForm) (error, logger.Severity) {
	res, err, sev := wholth.ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	sess := session.GetSession(r)

	if nil != err {
		return err, sev
	}

	if "" == form.Id {
		binds := make([]wholth.Bindable, 5)
		binds[0].Value = form.Exercise.Id
		binds[1].Value = sess.UserId
		binds[2].Value = form.TypeId
		binds[3].Value = form.Value
		binds[4].Value = form.CreatedAt

		err, sev = res.Bind2(binds)

		if nil != err {
			return err, sev
		}

		err, sev = res.Fetch("exercise_log_insert.sql")

		if nil != err {
			return err, sev
		}
	} else {
		binds := make([]wholth.Bindable, 5)
		binds[0].Value = form.Id
		binds[1].Value = sess.UserId
		binds[2].Value = form.TypeId
		binds[3].Value = form.Value

		if "" == form.CreatedAt {
			binds[4].IsNull = true
		} else {
			binds[4].Value = form.CreatedAt
		}

		err, sev = res.Bind2(binds)

		if nil != err {
			return err, sev
		}

		err, sev = res.Fetch("exercise_log_update.sql")

		if nil != err {
			return err, sev
		}

		form.Id = res.At(0, 0)
	}

	return nil, sev
}

func PostExerciseLog(w http.ResponseWriter, r *http.Request) {
	form := parseExerciseLogForm(r)
	isNew := "" == form.Exercise.Id

	page := PostExerciseLogPage{
		HtmlPage: route.DefaultHtmlPage(r),
		PostForm: form,
	}
	page.Meta.Title = form.Exercise.Title
	page.Status.Alias = "success"

	err, sev := saveExerciseLog(r, &form)
	if nil != err {
		container.Log(r, sev, "[PostExerciseLog]", err)
		page.Status.Alias = "error"
		page.Status.Message = err.Error()
		// util.TextResponse(w, http.StatusBadRequest, err.Error())
		// return
	} else {
		if isNew {
			page.Status.Message = "Создано!"
		} else {
			page.Status.Message = "Обновлено!"
		}
	}

	route.RenderHtmlTemplatesWithStatus(
		w,
		r,
		container.StatusCodeFromSeverity(sev),
		page,
		"templates/exercise_log/post/result.html",
	)
}
