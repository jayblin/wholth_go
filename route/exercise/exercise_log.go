package exercise

import (
	"fmt"
	"math"
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
	util.PaginatableList[wholth.ExerciseLog]
	util.Status
	ExerciseSearch struct {
		util.SearchForm
		util.PaginatableList[wholth.Exercise]
	}
	Groups   []string
	Map      map[string]MapElement
	From     wholth.DateTime
	To       wholth.DateTime
	Types    []wholth.ExerciseType
	PostForm PostExerciseLogForm
}

func ListExerciseLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	now := time.Now()

	// TODO adapt to user's timezone
	from := wholth.DateTimeCreate(q.Get("from_date"), q.Get("from_time"))
	_, fromErr := from.ToTime()
	if nil != fromErr {
		// from = time.Now().Add((time.Hour * 5) - (time.Hour * 48))
		from.UpdateFromTime(now.Add(time.Hour * -48).Truncate(time.Hour * 24))
	}

	to := wholth.DateTimeCreate(q.Get("to_date"), q.Get("to_time"))
	_, toErr := to.ToTime()
	if nil != toErr {
		// to = time.Now().Add(time.Hour * 5)
		to.UpdateFromTime(now.Add(time.Hour * 24).Truncate(time.Hour * 24))
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

			if len(value.CreatedAt.Date) > 0 {
				grp = value.CreatedAt.Date
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
	// TODO dedupe
	// @see grep DEDUP_TASK_1
	page.ExerciseSearch.SearchForm.Id = "exercise-history-aware-search"
	page.ExerciseSearch.SearchForm.Action = "/exercise-log/exercise-with-history"
	page.PaginatableList = list
	page.Types = types
	page.Groups = groups
	page.Map = mapped
	page.From = from
	page.To = to
	page.HtmlPage = route.DefaultHtmlPage(r)
	// page.PostForm = wholth.PostFoodsFormDefault()
	page.Meta.Title = fmt.Sprintf(
		"Логи упражнений // Страница %d",
		page.PaginatableList.Pagination.PageCurrent,
	)
	page.Meta.Description = ""
	page.PostForm = PostExerciseLogForm{
		Id: "",
	}

	history, err, sev := prepareExerciseHistory(r)

	if nil != err {
		container.Log(r, sev, "[ListExerciseLogs][prepareExerciseHistory]", err)
	} else {
		page.ExerciseSearch.Values = history
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
			"templates/utils/toggleable.html",

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

			"templates/exercise_log/get/exercise-with-history/search.html",
		)
	}
}

type PostExerciseLogForm struct {
	Id        string
	CreatedAt wholth.DateTime
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

	form := PostExerciseLogForm{
		Id:     r.PostForm.Get("id"),
		TypeId: r.PostForm.Get("type_id"),
		Value:  r.PostForm.Get("value"),
		Exercise: wholth.Exercise{
			Id: r.PostForm.Get("exercise_id"),
		},
	}

	form.CreatedAt = wholth.DateTimeCreate(
		r.PostForm.Get("created_at_date"),
		r.PostForm.Get("created_at_time"),
	)

	return form
}

func saveExerciseLog(r *http.Request, form *PostExerciseLogForm) (error, logger.Severity) {
	res, err, sev := wholth.ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	sess := session.GetSession(r)

	if nil != err {
		return err, sev
	}

	createdAt := form.CreatedAt.ToWholthFormat()
	_, createdAtErr := form.CreatedAt.ToTime()

	if "" == form.Id {
		binds := make([]wholth.Bindable, 5)
		binds[0].Value = form.Exercise.Id
		binds[1].Value = sess.UserId
		binds[2].Value = form.TypeId
		binds[3].Value = form.Value

		if nil == createdAtErr {
			binds[4].Value = createdAt
		} else {
			binds[4].IsNull = true
		}

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

		if "" == form.TypeId {
			binds[2].IsNull = true
		} else {
			binds[2].Value = form.TypeId
		}

		binds[3].Value = form.Value

		if nil == createdAtErr {
			binds[4].Value = createdAt
		} else {
			binds[4].IsNull = true
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

func deleteExerciseLog(r *http.Request, exerciseId string, userId string) (error, logger.Severity) {
	res, err, sev := wholth.ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		return err, sev
	}

	binds := make([]wholth.Bindable, 2)
	binds[0].Value = exerciseId
	binds[1].Value = userId

	err, sev = res.Bind2(binds)

	if nil != err {
		return err, sev
	}

	err, sev = res.Fetch("exercise_log_delete.sql")

	return err, sev
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

type BatchAction int

const (
	BatchDelete BatchAction = iota
	BatchPatch
)

func batchExerciseLog(action BatchAction, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var msg = ""
	switch action {
	case BatchPatch:
		{
			msg = "изменять"
			break
		}
	case BatchDelete:
		{
			msg = "удалять"
			break
		}
	}

	if !r.PostForm.Has("exercise_log") {
		route.RenderHtmlTemplatesWithStatus(
			w,
			r,
			400,
			fmt.Sprintf("Нечего %s!", msg),
			"templates/400/index.html",
		)
		return
	}

	ids := r.PostForm["exercise_log"]

	if len(ids) <= 0 {
		route.RenderHtmlTemplatesWithStatus(
			w,
			r,
			400,
			fmt.Sprintf("Нечего %s!", msg),
			"templates/400/index.html",
		)
		return
	}

	sess_v := r.Context().Value(session.SessionKey)
	sess := sess_v.(session.HttpSession)

	var successes = 0
	var errors = make([]string, 0)
	for _, id := range ids {
		var err error = nil

		switch action {
		case BatchPatch:
			{
				msg = "изменено"
				form := PostExerciseLogForm{
					Id:     id,
					Value:  r.PostForm.Get(fmt.Sprintf("value_%s", id)),
					TypeId: r.PostForm.Get(fmt.Sprintf("type_id_%s", id)),
				}
				form.CreatedAt = wholth.DateTimeCreate(
					r.PostForm.Get(fmt.Sprintf("created_at_date_%s", id)),
					r.PostForm.Get(fmt.Sprintf("created_at_time_%s", id)),
				)

				err, _ = saveExerciseLog(r, &form)
				break
			}
		case BatchDelete:
			{
				msg = "удалено"
				err, _ = deleteExerciseLog(r, id, sess.UserId)
				break
			}
		}

		if nil != err {
			errors = append(errors, fmt.Sprintf("id=%s; %s", id, err.Error()))
		} else {
			successes++
		}
	}

	result := make([][2]string, int(math.Min(1, float64(successes)))+len(errors))

	if successes > 0 {
		result[0] = [2]string{"success", fmt.Sprintf("Успешно %s %dшт.", msg, successes)}
	}

	for i, error := range errors {
		result[successes+i] = [2]string{"error", error}
	}

	route.RenderHtmlTemplates(
		w,
		r,
		result,
		"templates/exercise_log/batch/result.html",
	)
}

func BatchPatchExerciseLog(w http.ResponseWriter, r *http.Request) {
	batchExerciseLog(BatchPatch, w, r)
}

func BatchDeleteExerciseLog(w http.ResponseWriter, r *http.Request) {
	batchExerciseLog(BatchDelete, w, r)
}

type GetExercisesWithHistoryPage struct {
	route.HtmlPage
	util.SearchForm
	util.PaginatableList[wholth.Exercise]
}

func prepareExerciseHistory(r *http.Request) ([]wholth.Exercise, error, logger.Severity) {
	result := make([]wholth.Exercise, 0)

	// TODO add cache
	history, err, sev := wholth.FetchTodaysHistory(r, session.GetSession(r).UserId)

	if nil != err {
		return result, err, sev
	} else {
		deduped := make([]wholth.Exercise, 0)
		for _, el := range history {
			var is_already_inserted = false
			for _, d := range deduped {
				if d.Id == el.Exercise.Id {
					is_already_inserted = true
					break
				}
			}
			if !is_already_inserted {
				el.Exercise.PreferredType = el.Type
				deduped = append(deduped, el.Exercise)
			}
		}
		result = append(result, deduped...)
	}

	return result, nil, sev
}

func GetExercisesWithHistory(w http.ResponseWriter, r *http.Request) {
	// 1. get exercises
	pagination := util.QueryPaginationExtract(r.URL)

	list, err, sev := wholth.FetchExerciseList(r, pagination)

	if nil != err {
		container.Log(r, sev, "[GetExercisesWithHistory]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	page := GetExercisesWithHistoryPage{}
	page.PaginatableList = list
	page.HtmlPage = route.DefaultHtmlPage(r)
	page.Meta.Title = ""
	page.Meta.Description = ""
	// TODO dedupe
	// @see grep DEDUP_TASK_1
	page.SearchForm.Id = "exercise-history-aware-search"
	page.SearchForm.Action = "/exercise-log/exercise-with-history"

	history, err, sev := prepareExerciseHistory(r)

	if nil != err {
		container.Log(r, sev, "[GetExercisesWithHistory][prepareExerciseHistory]", err)
	} else {
		page.Values = append(page.Values, history...)
	}

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/exercise_log/get/exercise-with-history/as_subdoc.html",
		"templates/exercise_log/get/exercise-with-history/search.html",

		"templates/utils/search.html",
		"templates/utils/paginator.html",
	)
}
