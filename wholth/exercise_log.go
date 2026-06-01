package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"math"
	"net/http"
	"strconv"
	"time"
	"wholth_go/container"
	"wholth_go/logger"
	"wholth_go/util"
)

type ExerciseLog struct {
	util.ToggleableTrait

	Id        string
	Value     int64
	CreatedAt string
	Exercise  Exercise
	Type      ExerciseType
}

func (p ExerciseLog) EntityAlias() string {
	return p.Toggleable_Name()
}

func (p *ExerciseLog) Toggleable_Name() string {
	return "exercise_log"
}

func (p *ExerciseLog) Toggleable_Value() string {
	return p.Id
}

func FetchExerciseLogList(
	r *http.Request,
	userId string,
	from time.Time,
	to time.Time,
) (util.PaginatableList[ExerciseLog], error, logger.Severity) {
	pagination := util.QueryPaginationExtract(r.URL)

	list := util.PaginatableList[ExerciseLog]{}
	list.Count = 0
	list.PageCurrent = pagination.PageNumber + 1
	list.PageMax = 1

	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		return list, err, sev
	}

	format := DateFormat()

	binds := make([]Bindable, 7)

	binds[0].Value = userId

	binds[1].IsNull = true // id

	if "" == pagination.Q {
		binds[2].IsNull = true
	} else {
		binds[2].Value = pagination.Q
	}

	binds[3].Value = from.Format(format)
	binds[4].Value = to.Format(format)

	binds[5].Value = strconv.FormatUint(pagination.Limit, 10)
	binds[6].Value = strconv.FormatUint(pagination.PageNumber, 10)

	err, sev = res.Bind2(binds)

	if nil != err {
		return list, err, sev
	}

	err, sev = res.Fetch("exercise_log_select.sql")

	if nil != err {
		return list, err, sev
	}

	sz := res.RowCount()

	if 0 == sz {
		return list, err, sev
	}

	cnt, _ := strconv.ParseUint(res.At(0, 0), 10, 64)
	list.Pagination.Count = cnt
	list.Pagination.PageMax = uint64(math.Ceil(float64(cnt) / float64(pagination.Limit)))
	list.Values = make([]ExerciseLog, sz-1)

	for i := range sz - 1 {
		j := uint(i + 1)
		value, _ := strconv.ParseInt(res.At(j, 1), 10, 64)
		list.Values[i] = ExerciseLog{
			Id:        res.At(j, 0),
			Value:     value,
			CreatedAt: res.At(j, 2),
			Exercise: Exercise{
				Id:    res.At(j, 3),
				Title: res.At(j, 4),
			},
			Type: ExerciseType{
				Id:    res.At(j, 5),
				Alias: res.At(j, 6),
				Unit:  res.At(j, 7),
			},
		}
	}

	return list, err, sev
}
