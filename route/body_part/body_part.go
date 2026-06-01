package body_part

import (
	"net/http"

	"wholth_go/cache"
	"wholth_go/container"
	"wholth_go/route"
	"wholth_go/util"
	"wholth_go/wholth"
)

type ListBodyPartPage struct {
	route.HtmlPage
	util.PaginatableList[wholth.BodyPart]
	PostForm wholth.PostBodyPartForm
}

func ListBodyParts(w http.ResponseWriter, r *http.Request) {
	// q := util.QueryPaginationExtract(r.URL)
	// container.AddBreadcrumb("ListBodyParts")

	list, err, sev := wholth.FetchBodyPartList(r)

	if nil != err {
		// logger.Log(sev, err.Error())
		container.Log(r, sev, "[ListBodyParts]", err)
		util.TextResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	page := ListBodyPartPage{}
	page.PaginatableList = list
	page.HtmlPage = route.DefaultHtmlPage(r)
	// page.PostForm = wholth.PostFoodsFormDefault()
	page.Meta.Title = "body_part"
	page.Meta.Description = ""

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/index.html",

		"templates/body_part/tree.html",
		"templates/body_part/get/content.html",
		"templates/body_part/post/form.html",
	)
}

func PostBodyPart(w http.ResponseWriter, r *http.Request) {
	form, err, sev := wholth.SaveBodyPart(r)

	if nil != err {
		container.Log(r, sev, "[PostBodyPart]", err)
		util.TextResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	cache.DeleteByTag("body_part")

	page := ListBodyPartPage{}
	page.PostForm = form
	page.HtmlPage = route.DefaultHtmlPage(r)
	page.Meta.Title = "Часьти тела"
	page.Meta.Description = ""

	list, err, sev := wholth.FetchBodyPartList(r)

	if nil != err {
		container.Log(r, sev, "[PostBodyPart]", err)
	} else {
		page.Values = list.Values
	}

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/body_part/post/result.html",
		"templates/body_part/tree.html",
	)
}
