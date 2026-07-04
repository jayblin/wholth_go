package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"net/http"
	"strconv"
	"wholth_go/cache"
	"wholth_go/container"
	"wholth_go/logger"
	"wholth_go/util"
)

type BodyPart struct {
	util.ToggleableTrait
	Id          string
	Title       string
	Description string
	Lft         int
	Rgt         int
	ChildCount  int
	Lvl         int

	// util.Status
	// Value   string
}

func (f BodyPart) EntityAlias() string {
	return "body_part"
}

func FetchBodyPartList(r *http.Request) (util.PaginatableList[BodyPart], error, logger.Severity) {
	result := util.PaginatableList[BodyPart]{}
	cached, err := cache.GetV2[util.PaginatableList[BodyPart]]("FetchBodyPartList")

	if nil != err {
		return result, err, logger.ALERT
	}

	if nil != cached {
		values := make([]BodyPart, len(cached.Values))
		copy(values, cached.Values)
		result.Values = values
		result.Pagination = cached.Pagination
		return result, nil, logger.DEBUG
	}

	list := util.PaginatableList[BodyPart]{}
	list.Count = 0
	list.PageCurrent = 0
	list.PageMax = 0

	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		// logger.Error("[FetchBodyPartList][1]", err)
		return list, err, logger.EMERGENCY
	}

	binds := make([]Bindable, 0)

	err, sev = res.Bind2(binds)

	if nil != err {
		// logger.Error("[FetchBodyPartList][2]", err)
		return list, err, sev
	}

	err, sev = res.Fetch("body_part_select.sql")

	if nil != err {
		// logger.Error("[FetchBodyPartList][3]", err)
		return list, err, sev
	}

	sz := res.RowCount()

	if sz == 0 {
		list.Count = 0
		return list, nil, logger.DEBUG
	}
	list.Count = sz
	list.Values = make([]BodyPart, sz)

	var lvl = 1
	var prev *BodyPart = nil
	for i := range uint(sz) {
		lft, _ := strconv.ParseInt(res.At(i, 3), 10, 64)
		rgt, _ := strconv.ParseInt(res.At(i, 4), 10, 64)
		cnt := int((rgt - lft - 1) / 2)

		list.Values[i] = BodyPart{
			Id:          res.At(i, 0),
			Title:       res.At(i, 1),
			Description: res.At(i, 2),
			Lft:         int(lft),
			Rgt:         int(rgt),
			ChildCount:  cnt,
			Lvl:         0,
		}
		cur := &list.Values[i]

		if nil != prev {
			diff := cur.Lft - prev.Rgt - 1
			if diff >= 0 {
				lvl = lvl - diff
			} else {
				lvl++
			}
		}

		cur.Lvl = lvl

		prev = cur
	}

	cache.SetV2("FetchBodyPartList", list, "body_part")
	values := make([]BodyPart, len(list.Values))
	copy(values, list.Values)
	result.Values = values
	result.Pagination = list.Pagination

	return result, nil, logger.DEBUG
}

type PostBodyPartForm struct {
	util.Status
	BodyPart BodyPart
	ParentId string
}

func PostBodyPartFormFromRequest(r *http.Request) PostBodyPartForm {
	r.ParseForm()
	result := PostBodyPartForm{
		BodyPart: BodyPart{
			Id:          r.PostForm.Get("id"),
			Title:       r.PostForm.Get("title"),
			Description: r.PostForm.Get("description"),
		},
		ParentId: r.PostForm.Get("parent_id"),
	}

	return result
}

func SaveBodyPart(r *http.Request) (PostBodyPartForm, error, logger.Severity) {
	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		return PostBodyPartForm{}, err, sev
	}

	form := PostBodyPartFormFromRequest(r)

	binds := make([]Bindable, 3)
	binds[0].Value = form.BodyPart.Title
	binds[1].Value = form.BodyPart.Description
	binds[2].Value = form.ParentId

	err, sev = res.Bind2(binds)

	if nil != err {
		return form, err, sev
	}

	err, sev = res.Fetch("body_part_insert.sql")

	if nil != err {
		return form, err, sev
	}

	sz := res.RowCount()

	if sz == 0 {
		return form, err, logger.DEBUG
	}

	form.BodyPart.Id = res.At(0, 0)

	return form, nil, logger.DEBUG
}
