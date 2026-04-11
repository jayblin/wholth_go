package util

import (
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
	Toggleable_Name()    string
	Toggleable_Value()   string
}

// func
