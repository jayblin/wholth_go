package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"wholth_go/container"
	"wholth_go/logger"
	"wholth_go/util"
)

type ExerciseType struct {
	Id    string
	Alias string
	Unit  string
}

type Exercise struct {
	Id            string
	Title         string
	Description   string
	PreferredType ExerciseType
	BodyParts     []BodyPart
}

func (f Exercise) EntityAlias() string {
	return "exercise"
}

func FetchExerciseTypeList(r *http.Request) ([]ExerciseType, error, logger.Severity) {
	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		container.Log(r, sev, "[FetchExerciseTypeList][0]", err)
		return nil, err, sev
	}

	err, sev = res.Fetch("exercise_type_select.sql")

	if nil != err {
		container.Log(r, sev, "[FetchExerciseTypeList][1]", err)
		return nil, err, sev
	}

	sz := res.RowCount()

	if sz == 0 {
		return nil, nil, sev
	}

	arr := make([]ExerciseType, sz)

	for i := range uint(sz) {
		arr[i] = ExerciseType{
			Id:    res.At(i, 0),
			Alias: res.At(i, 1),
			Unit:  res.At(i, 2),
		}
	}

	return arr, nil, sev
}

func mapBodyPartIds(exercise *Exercise, rawBodyPartIds string) {
	bodyPartIds := strings.SplitN(rawBodyPartIds, ",", 100)
	exercise.BodyParts = make([]BodyPart, len(bodyPartIds))

	for i, id := range bodyPartIds {
		exercise.BodyParts[i].Id = id
	}
}

func FetchExercise(r *http.Request, id string) (Exercise, error, logger.Severity) {
	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		return Exercise{}, err, sev
	}

	binds := make([]Bindable, 4)

	// id
	binds[0].Value = id
	// title
	binds[1].IsNull = true
	// limit
	binds[2].Value = "1"
	// page_num
	binds[3].Value = "0"

	err, sev = res.Bind2(binds)

	if nil != err {
		return Exercise{}, err, sev
	}

	err, sev = res.Fetch("exercise_select.sql")

	if nil != err {
		return Exercise{}, err, sev
	}

	sz := res.RowCount()

	if sz < 2 {
		return Exercise{}, nil, sev
	}

	exercise := Exercise{
		Id:          res.At(1, 0),
		Title:       res.At(1, 1),
		Description: res.At(1, 2),
		PreferredType: ExerciseType{
			Id:    res.At(1, 3),
			Alias: res.At(1, 4),
			Unit:  res.At(1, 5),
		},
	}

	 mapBodyPartIds(&exercise, res.At(1, 6))

	return exercise, nil, sev
}

func FetchExerciseList(
	r *http.Request,
	pagination util.QueryPagination,
) (util.PaginatableList[Exercise], error, logger.Severity) {
	list := util.PaginatableList[Exercise]{}
	list.Count = 0
	list.PageCurrent = pagination.PageNumber + 1
	list.PageMax = 1

	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		container.Log(r, sev, "[FetchExerciseList][0]", err)
		return list, err, logger.ALERT
	}

	binds := make([]Bindable, 4)

	// id
	binds[0].IsNull = true

	// title
	if "" == pagination.Q {
		binds[1].IsNull = true
	} else {
		binds[1].Value = "{title}:" + pagination.Q
	}

	binds[2].Value = strconv.FormatUint(pagination.Limit, 10)

	binds[3].Value = strconv.FormatUint(pagination.PageNumber, 10)

	err, sev = res.Bind2(binds)

	if nil != err {
		return list, err, logger.ALERT
	}

	err, sev = res.Fetch("exercise_select.sql")

	if nil != err {
		return list, err, sev
	}

	sz := res.RowCount()

	if sz == 0 {
		return list, nil, logger.DEBUG
	}

	cnt, _ := strconv.ParseUint(res.At(0, 0), 10, 64)
	list.Pagination.Count = cnt
	list.Pagination.PageMax = uint64(math.Ceil(float64(cnt) / float64(pagination.Limit)))
	list.Values = make([]Exercise, sz-1)

	for i := range sz - 1 {
		j := uint(i + 1)
		list.Values[i] = Exercise{
			Id:          res.At(j, 0),
			Title:       res.At(j, 1),
			Description: res.At(j, 2),
			PreferredType: ExerciseType{
				Id:    res.At(j, 3),
				Alias: res.At(j, 4),
				Unit:  res.At(j, 5),
			},
		}

		 mapBodyPartIds(&list.Values[i], res.At(1, 6))
	}

	bodyParts, err, sev := FetchBodyPartList(r)

	if nil != err {
		return list, err, sev
	}

	for i := range list.Values {
		for j := range bodyParts.Values {
			for k := range list.Values[i].BodyParts {
				if list.Values[i].BodyParts[k].Id == bodyParts.Values[j].Id {
					list.Values[i].BodyParts[k] = bodyParts.Values[j]
					break
				}
			}
		}
	}

	return list, nil, logger.DEBUG
}

type PostExerciseForm struct {
	Exercise     Exercise
	BodyPartsIds []string
}

func syncExerciseBodyPart(r *http.Request, exerciseId string, bodyPartId string) (error, logger.Severity) {
	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		return err, sev
	}

	binds := make([]Bindable, 2)
	binds[0].Value = exerciseId
	binds[1].Value = bodyPartId

	err, sev = res.Bind2(binds)

	if nil != err {
		return err, sev
	}

	return res.Fetch("exercise_body_part_insert.sql")
}

func syncExerciseBodyParts(r *http.Request, form *PostExerciseForm) []error {
	if len(form.BodyPartsIds) <= 0 {
		return nil
	}

	errs := make([]error, len(form.BodyPartsIds))

	for i, id := range form.BodyPartsIds {
		err, _ := syncExerciseBodyPart(r, form.Exercise.Id, id)
		errs[i] = err
	}

	return errs
}

func SaveExercise(r *http.Request, form *PostExerciseForm) (error, logger.Severity) {
	res, err, sev := ExecStmtResultNew()
	defer res.ContainedDelete(container.Instance(r))

	if nil != err {
		return err, sev
	}

	if "" != form.Exercise.Id {
		binds := make([]Bindable, 4)
		binds[0].Value = form.Exercise.Id
		binds[1].Value = form.Exercise.Title
		binds[2].Value = form.Exercise.Description
		binds[3].Value = form.Exercise.PreferredType.Id

		err, sev = res.Bind2(binds)

		if nil != err {
			return err, sev
		}

		err, sev = res.Fetch("exercise_update.sql")

		if nil != err {
			return err, sev
		}
	} else {
		binds := make([]Bindable, 3)
		binds[0].Value = form.Exercise.Title
		binds[1].Value = form.Exercise.Description
		binds[2].Value = form.Exercise.PreferredType.Id

		err, sev = res.Bind2(binds)

		if nil != err {
			return err, sev
		}

		err, sev = res.Fetch("exercise_insert.sql")

		if nil != err {
			return err, sev
		}

		form.Exercise.Id = res.At(0, 0)
	}

	return errors.Join(syncExerciseBodyParts(r, form)...), logger.DEBUG
}
