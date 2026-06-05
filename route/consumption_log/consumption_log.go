package consumption_log

import (
	"fmt"
	"math"
	"net/http"
	"time"

	// "wholth_go/logger"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/session"
	"wholth_go/util"
	"wholth_go/wholth"
)

type DummyFoodList struct {
	util.PaginatableList[wholth.Food]
}

type MapElement struct {
	NutrientAmountSum float64
	Values            []wholth.ConsumptionLog
}

type ListConsumptionLogsPage struct {
	route.HtmlPage
	util.Pagination
	util.EntityAliasAware[wholth.ConsumptionLog]
	Groups        []string
	Map           map[string]MapElement
	Q             string
	PostForm      wholth.ConsumptionLogPostForm
	ConsumedFrom  string
	ConsumedTo    string
	DummyFoodList DummyFoodList
}

func ListConsumptionLogs(w http.ResponseWriter, r *http.Request) {
	page, pageErr := wholth.ConsumptionLogPageNew(50)

	defer page.Close()

	if nil != pageErr {
		logger.Alert(pageErr.Error())
		w.Write([]byte(pageErr.Error()))
		return
	}

	sess_v := r.Context().Value(session.SessionKey)
	sess := sess_v.(session.HttpSession)
	page.SetUserId(sess.UserId)

	format := wholth.DateFormat()

	q := r.URL.Query()
	var from, fromErr = time.Parse(
		format,
		q.Get("consumed_from"))

	if nil != fromErr {
		// TODO adapt to user's timezone
		from = time.Now().Add(time.Hour * 5).Add(-48 * time.Hour)
	}

	var to, toErr = time.Parse(
		format,
		q.Get("consumed_to"))

	if nil != toErr {
		// TODO adapt to user's timezone
		to = time.Now().Add(time.Hour * 5)
	}

	page.SetPeriod(from, to)

	page.Fetch()

	size := page.Size()

	mapped := make(map[string]MapElement)
	groups := make([]string, 0)

	if size > 0 {
		for i := size; i > 0; {
			i--
			value := page.At(i)

			// 1234567890123456789
			// 2025-12-14T23:48:53
			grp := value.ConsumedAt[0:10]

			// value.ConsumedAt = value.ConsumedAt[11:]

			entry, ok := mapped[grp]

			if !ok {
				entry = MapElement{
					Values:            make([]wholth.ConsumptionLog, 0),
					NutrientAmountSum: 0,
				}
				mapped[grp] = entry
				groups = append(groups, grp)
			}

			if 0.0 != value.NutrientAmount {
				entry.NutrientAmountSum += value.NutrientAmount
				entry.NutrientAmountSum = math.RoundToEven(entry.NutrientAmountSum)
			}

			entry.Values = append(entry.Values, value)

			mapped[grp] = entry
		}
	}

	// tz, tzErr := time.LoadLocation("Asia/Yekaterinburg")
	//
	// if nil != tzErr {
	// 	// todo log learn about panic
	// 	return
	// }

	htmlPage := ListConsumptionLogsPage{
		HtmlPage: route.DefaultHtmlPage(r),
		Groups:   groups,
		PostForm: wholth.ConsumptionLogPostForm{
			Mass:       "",
			// TODO adapt to user's timezone
			ConsumedAt: time.Now().Add(time.Hour * 5).Format(format),
		},
		ConsumedFrom: from.Format(format),
		ConsumedTo:   to.Format(format),
		Map:          mapped,
		Pagination:   page.Pagination(),
	}
	htmlPage.Meta.Title = "Логи"
	htmlPage.Meta.Description = "Логи поедания"

	as_subdoc := q.Get("as_subdoc")

	if "" != as_subdoc {
		route.RenderHtmlTemplates(
			w,
			r,
			htmlPage,
			"templates/consumption_log/get/as_subdoc.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",
			"templates/utils/toggleable.html",

			"templates/consumption_log/get/form.html",
		)
	} else {
		route.RenderHtmlTemplates(
			w,
			r,
			htmlPage,
			"templates/index.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",
			"templates/utils/toggleable.html",

			"templates/consumption_log/get/content.html",
			"templates/consumption_log/get/form.html",
			"templates/consumption_log/post/form.html",
			"templates/consumption_log/post/result.html",

			"templates/food/get/as_subdoc/content.html",
			"templates/food/get/suggestion.html",
		)
	}
}

type PostConsumptionLogPage struct {
	route.HtmlPage
	PostForm wholth.ConsumptionLogPostForm
}

func PostConsumptionLog(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	format := wholth.DateFormat()

	form := wholth.ConsumptionLogPostForm{
		Id:            r.PostForm.Get("id"),
		FoodId:        r.PostForm.Get("food_id"),
		FoodTitle:     r.PostForm.Get("q"),
		Mass:          r.PostForm.Get("mass"),
		ConsumedAt:    "",
		ResultStatus:  "",
		ResultMessage: "",
	}

	consumedAt, consumedAtErr := time.Parse(
		format,
		r.PostForm.Get("consumed_at"))

	if nil != consumedAtErr {
		page := PostConsumptionLogPage{}
		page.PostForm.ResultStatus = "error"
		page.PostForm.ResultMessage = "Невалидная дата поедания!"

		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/consumption_log/post/result.html",
		)

		return
	}

	form.ConsumedAt = consumedAt.Format(format)

	sess_v := r.Context().Value(session.SessionKey)
	sess := sess_v.(session.HttpSession)

	status, err := wholth.SaveConsumptionLog(&form, sess.UserId)

	page := PostConsumptionLogPage{
		route.DefaultHtmlPage(r),
		form,
	}
	page.Meta.Title = "Логи"
	page.Meta.Description = "Логи поедания"

	if nil != err {
		page.PostForm.ResultStatus = status
		page.PostForm.ResultMessage = err.Error()
	} else {
		page.PostForm.ResultStatus = status
		page.PostForm.ResultMessage = "Успешно сохранено!"
	}

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/consumption_log/post/result.html",
	)
}

type BatchAction int

const (
	BatchDelete BatchAction = iota
	BatchPatch
)

func batchConsumptionLog(action BatchAction, w http.ResponseWriter, r *http.Request) {
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

	if !r.PostForm.Has("consumption_log") {
		route.RenderHtmlTemplatesWithStatus(
			w,
			r,
			400,
			fmt.Sprintf("Нечего %s!", msg),
			"templates/400/index.html",
		)
		return
	}

	ids := r.PostForm["consumption_log"]

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
				mass := r.PostForm.Get(fmt.Sprintf("mass_%s", id))
				consumedAt := r.PostForm.Get(fmt.Sprintf("consumed_at_%s", id))
				err = wholth.UpdateConsumptionLog(r, id, mass, consumedAt, sess.UserId)
				break
			}
		case BatchDelete:
			{
				msg = "удалено"
				err = wholth.DeleteConsumptionLog(r, id, sess.UserId)
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
		"templates/consumption_log/batch/result.html",
	)
}

func BatchPatchConsumptionLog(w http.ResponseWriter, r *http.Request) {
	batchConsumptionLog(BatchPatch, w, r)
}

func BatchDeleteConsumptionLog(w http.ResponseWriter, r *http.Request) {
	batchConsumptionLog(BatchDelete, w, r)
}
