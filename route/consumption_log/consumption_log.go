package consumption_log

import (
	"net/http"
	"time"

	// "wholth_go/logger"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/session"
	"wholth_go/util"
	"wholth_go/wholth"
)

type GetListPage struct {
	route.HtmlPage
	// Values       []wholth.ConsumptionLog
	Groups       []string
	Values       map[string][]wholth.ConsumptionLog
	Pagination   util.Pagination
	PostForm     wholth.ConsumptionLogPostForm
	ConsumedFrom string
	ConsumedTo   string
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
		from = time.Now().Add(-48 * time.Hour)
	}

	var to, toErr = time.Parse(
		format,
		q.Get("consumed_to"))

	if nil != toErr {
		to = time.Now()
	}

	page.SetPeriod(from, to)

	page.Fetch()

	size := page.Size()

	mapped := make(map[string][]wholth.ConsumptionLog)
	groups := make([]string, 0)

	for i := (size - 1); i > 0; i-- {
		value := page.At(i)

		// 1234567890123456789
		// 2025-12-14T23:48:53
		grp := value.ConsumedAt[0:10]
		values, ok := mapped[grp]

		value.ConsumedAt = value.ConsumedAt[11:]

		if ok {
			mapped[grp] = append(values, value)
		} else {
			values = make([]wholth.ConsumptionLog, 1)
			values[0] = value
			mapped[grp] = values
			groups = append(groups, grp)
		}
	}

	// tz, tzErr := time.LoadLocation("Asia/Yekaterinburg")
	//
	// if nil != tzErr {
	// 	// todo log learn about panic
	// 	return
	// }

	htmlPage := GetListPage{
		route.DefaultHtmlPage(r),
		groups,
		mapped,
		page.Pagination(),
		wholth.ConsumptionLogPostForm{
			Mass:       "0",
			ConsumedAt: time.Now().Format(format),
		},
		from.Format(format),
		to.Format(format),
	}
	htmlPage.Meta.Title = "Логи"
	htmlPage.Meta.Description = "Логи поедания"

	as_subdoc := q.Get("as_subdoc")

	if "" != as_subdoc {
		route.RenderHtmlTemplates(
			w,
			r,
			htmlPage,
			"templates/consumption_log/get/form.html",
			"templates/utils/paginator.html",
		)
	} else {
		route.RenderHtmlTemplates(
			w,
			r,
			htmlPage,
			"templates/index.html",
			"templates/consumption_log/get/content.html",
			"templates/consumption_log/get/form.html",
			"templates/consumption_log/post/form.html",
			"templates/food/get/suggestion_list.html",
			"templates/utils/paginator.html",
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
		FoodTitle:     r.PostForm.Get("food_q"),
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
