package exercise

import (
	"fmt"
	"net/http"

	// "slices"

	"wholth_go/container"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/util"
	"wholth_go/wholth"
)

type ListExercisePage struct {
	route.HtmlPage
	util.PaginatableList[wholth.Exercise]
}

func ListExercises(w http.ResponseWriter, r *http.Request) {
	pagination := util.QueryPaginationExtract(r.URL)

	list, err, sev := wholth.FetchExerciseList(r, pagination)

	if nil != err {
		container.Log(r, sev, "[ListExercises]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	page := ListExercisePage{}
	page.PaginatableList = list
	page.HtmlPage = route.DefaultHtmlPage(r)
	// page.PostForm = wholth.PostFoodsFormDefault()
	page.Meta.Title = fmt.Sprintf(
		"Упражнения // Страница %d",
		page.PaginatableList.Pagination.PageCurrent,
	)
	page.Meta.Description = ""

	q := r.URL.Query()

	if "" != q.Get("as_subdoc") {
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/exercise/get/as_subdoc.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",

			"templates/exercise/get/search.html",
		)
	} else if "" != q.Get("as_radio") {
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/exercise/get/as_radio.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",
		)
	} else {
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/index.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",

			"templates/exercise/get/search.html",
			"templates/exercise/get/content.html",
		)
	}
}

type PostExercisePage struct {
	route.HtmlPage
	util.Status
	PostForm      wholth.PostExerciseForm
	BodyParts     util.PaginatableList[wholth.BodyPart]
	ExerciseTypes []wholth.ExerciseType
}

// used in templates
func (r PostExercisePage) IsExerciseTypeSelected(exerciseTypeId string) bool {
	return r.PostForm.Exercise.PreferredType.Id == exerciseTypeId
}

func GetExerciseForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	form := wholth.PostExerciseForm{}

	if "add" != id {
		entity, err, sev := wholth.FetchExercise(r, id)

		if nil != err {
			container.Log(r, sev, "[GetExerciseForm]", err)
			util.TextResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if entity.Id == "" {
			route.RenderHtmlTemplatesWithStatus(
				w,
				r,
				http.StatusNotFound,
				"",
				"templates/index.html",
				"templates/404/content.html",
			)
			return
		}

		form.Exercise = entity
	}

	bodyParts, err, sev := wholth.FetchBodyPartList(r)

	if nil != err {
		container.Log(r, sev, "[GetExerciseForm]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for i := range bodyParts.Values {
		for j := range form.Exercise.BodyParts {
			if bodyParts.Values[i].Id == form.Exercise.BodyParts[j].Id {
				bodyParts.Values[i].Checked = true
				continue
			}
		}
	}

	types, err, sev := wholth.FetchExerciseTypeList(r)

	if nil != err {
		container.Log(r, sev, "[GetExerciseForm]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	page := PostExercisePage{
		HtmlPage:      route.DefaultHtmlPage(r),
		PostForm:      form,
		BodyParts:     bodyParts,
		ExerciseTypes: types,
	}
	page.Meta.Title = form.Exercise.Title

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/index.html",

		"templates/body_part/tree.html",

		"templates/exercise/_id/get/content.html",
		"templates/exercise/post/form.html",
	)
}

func parseForm(r *http.Request) wholth.PostExerciseForm {
	r.ParseForm()
	form := wholth.PostExerciseForm{
		Exercise: wholth.Exercise{
			Id:          r.PostForm.Get("id"),
			Title:       r.PostForm.Get("title"),
			Description: r.PostForm.Get("description"),
			PreferredType: wholth.ExerciseType{
				Id: r.PostForm.Get("preffered_type_id"),
			},
		},
	}
	form.BodyPartsIds = r.PostForm["body_parts"]

	return form
}

func PostExercise(w http.ResponseWriter, r *http.Request) {
	form := parseForm(r)
	isNew := "" == form.Exercise.Id
	err, sev := wholth.SaveExercise(r, &form)

	if nil != err {
		container.Log(r, sev, "[PostExercise]", err)
		util.TextResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	bodyParts, err, sev := wholth.FetchBodyPartList(r)

	if nil != err {
		container.Log(r, sev, "[PostExercise]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	types, err, sev := wholth.FetchExerciseTypeList(r)

	if nil != err {
		container.Log(r, logger.ERROR, "[PostExercise]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	page := PostExercisePage{
		HtmlPage:      route.DefaultHtmlPage(r),
		PostForm:      form,
		BodyParts:     bodyParts,
		ExerciseTypes: types,
	}
	page.Meta.Title = form.Exercise.Title

	page.Status.Alias = "success"
	if isNew {
		page.Status.Message = "Создано!"
	} else {
		page.Status.Message = "Обновлено!"
	}

	// w.Header().Add("Location", "/exercise/"+form.Exercise.Id)
	route.RenderHtmlTemplatesWithStatus(
		w,
		r,
		http.StatusOK,
		page,
		"templates/exercise/post/result.html",
	)
}
