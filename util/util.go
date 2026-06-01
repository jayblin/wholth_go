package util

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func ArrayFilter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

type Pagination struct {
	PageCurrent uint64
	PageMax     uint64
	Count       uint64
}

type AliasableEntity interface {
	EntityAlias() string
}

type EntityAliasAware[T AliasableEntity] struct {
	Dummy T
}

func (p EntityAliasAware[T]) EntityAlias() string {
	return p.Dummy.EntityAlias()
}

type PaginatableList[T AliasableEntity] struct {
	EntityAliasAware[T]
	Pagination
	Values []T
	Q      string
}

type Status struct {
	Alias   string
	Message string
}

func ModifyQueryForSearch(q string) string {
	q = strings.ToLower(q)

	if "" == q {
		return q
	}

	split := strings.SplitN(q, ",", 20)

	for i, subq := range split {
		subq = strings.Trim(subq, " \t\v\r\n")

		if "" == subq {
			continue

		}

		if '*' != subq[len(subq)-1] {
			split[i] = subq + "*"
		} else {
			split[i] = subq
		}
	}

	res := strings.Join(split, ",")

	return res
}

// type ToggleableData struct {
// 	Checked bool
// 	// Name    string
// 	// Value   string
// }

type Toggleable interface {
	Toggleable_Checked() bool
	Toggleable_Name() string
	Toggleable_Value() string
}

type ToggleableTrait struct {
	Checked bool
}

func (p *ToggleableTrait) Toggleable_Checked() bool {
	return p.Checked
}

func (p *ToggleableTrait) Toggleable_Name() string {
	return "[TOGGLEABLE_TRAIT_DUMMY_NAME]"
}

func (p *ToggleableTrait) Toggleable_Value() string {
	return "[TOGGLEABLE_TRAIT_DUMMY_ID]"
}

type QueryPagination struct {
	Limit      uint64
	PageNumber uint64
	Q          string
}

// TODO redo into PaginatableList??
func QueryPaginationExtract(u *url.URL) QueryPagination {
	result := QueryPagination{}

	q := u.Query()
	// var limit, limit_err = strconv.Atoi(q.Get("limit"))
	var limit, limit_err = strconv.ParseUint(q.Get("limit"), 10, 64)

	if nil != limit_err || 0 == limit {
		result.Limit = 20
	} else if limit > 100 {
		result.Limit = 100
	} else {
		result.Limit = limit
	}

	// page_number, page_number_err := strconv.Atoi(q.Get("page_number"))
	page_number, page_number_err := strconv.ParseUint(q.Get("page_number"), 10, 64)

	if nil != page_number_err || 0 == page_number {
		result.PageNumber = 0
	} else {
		// check for int overflow
		result.PageNumber = page_number - 1
	}

	result.Q = ModifyQueryForSearch(q.Get("q"))

	return result
}

// todo move to route namespace
func TextResponse(w http.ResponseWriter, status int, text string) {
	w.Header().Add("content-type", "text/plain;charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(text))
}
