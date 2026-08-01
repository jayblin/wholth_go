package wholth

// #cgo CFLAGS: -I${SRCDIR}/../wholth_lib/include
// #define WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #include "wholth/c/exec_stmt.h"
import "C"

import (
	"errors"
	"regexp"
	"runtime"
	"strconv"
	"time"
	"unsafe"

	"wholth_go/container"
	"wholth_go/logger"
	"wholth_go/util"
)

var DEFAULT_LOCALE_ID string = "2"

func toStr(sv C.wholth_StringView) string {
	if nil == sv.data || 0 == sv.size {
		return ""
	}
	return C.GoStringN(sv.data, C.int(sv.size)) // uh oh size cast
}

func toStrView(s string) C.wholth_StringView {
	c_str := C.CString(s)
	return C.wholth_StringView{
		c_str,
		C.ulong(len(s)),
	}
}

func DateFormat() string {
	return "2006-01-02T15:04:05"
}

func Setup() {
	db_path := C.CString("./wholth.db")
	ctx := C.wholth_AppSetupArgs{db_path}
	setup_err := C.wholth_app_setup(&ctx)

	if !C.wholth_error_ok(&setup_err) {
		message := toStr(setup_err.message)
		// logger.Emergency(message)
		panic(message)
	}
}

func SetPasswordEncryptionSecret(secret string) {
	C.wholth_app_password_encryption_secret(toStrView(secret))
}

type PostFoodsForm struct {
	Food        Food
	RecipeStep  RecipeStep
	Ingredients util.PaginatableList[Ingredient]
	Nutrients   util.PaginatableList[FoodNutrient]
	Target      string
	Action      string
	// ResultStatus  string
	// ResultMessage string
	util.Status
}

func PostFoodsFormDefault() PostFoodsForm {
	result := PostFoodsForm{}

	return result
}

func saveBasics(buf *C.wholth_Buffer, form *PostFoodsForm) error {

	food := C.wholth_entity_food_init()
	food.id = toStrView(form.Food.Id)
	food.title = toStrView(form.Food.Title)

	// if (svs[2].size > 0) {
	// 	food.description = svs[2]
	// }

	var err = C.wholth_Error_OK

	if "" != form.Food.Id {
		err = C.wholth_em_food_update(&food, toStrView(DEFAULT_LOCALE_ID), buf)
	} else {
		err = C.wholth_em_food_insert(&food, toStrView(DEFAULT_LOCALE_ID), buf)
	}

	if !C.wholth_error_ok(&err) {
		return errors.New("Ошибка сохранения общей инф-ии: " + toStr(err.message))
	}

	// copying food.id from scratch-buffer, cuz scratch may be modified!
	form.Food.Id = toStr(food.id)

	return nil
}

func saveNutrients(buf *C.wholth_Buffer, form *PostFoodsForm) error {
	food := C.wholth_entity_food_init()
	food.id = toStrView(form.Food.Id)

	var ok = true
	for i := range form.Nutrients.Values {
		wnut := C.wholth_entity_nutrient_init()
		wnut.id = toStrView(form.Nutrients.Values[i].Id)
		wnut.value = toStrView(form.Nutrients.Values[i].Value)

		var err = C.wholth_em_food_nutrient_upsert(&food, &wnut, buf)

		if !C.wholth_error_ok(&err) {
			form.Nutrients.Values[i].Status.Alias = "error"
			form.Nutrients.Values[i].Status.Message = toStr(err.message)
			ok = false
		}
	}

	if !ok {
		return errors.New("Не удалось сохранить нутриенты!")
	}

	return nil
}

type ExecStmtResult struct {
	Handle        *C.wholth_exec_stmt_Result
	Pinner        *runtime.Pinner
	FirstBindable *C.wholth_exec_stmt_Bindable
	BindableCount uint
	// _Error        error
}

func ExecStmtResultNew() (ExecStmtResult, error, logger.Severity) {
	result := ExecStmtResult{
		Handle:        nil,
		Pinner:        &runtime.Pinner{},
		FirstBindable: nil,
		BindableCount: 0,
	}
	result.Pinner.Pin(result.Handle)
	err := C.wholth_exec_stmt_Result_new(&result.Handle)

	if !C.wholth_error_ok(&err) {
		msg := toStr(err.message)
		// logger.Alert(msg)
		return result, errors.New(msg), logger.ALERT
	}

	return result, nil, logger.DEBUG
}

func _ExecStmtResultDelete(r *ExecStmtResult) (error, logger.Severity) {
	defer r.Pinner.Unpin()

	err := C.wholth_exec_stmt_Result_del(r.Handle)

	if !C.wholth_error_ok(&err) {
		msg := toStr(err.message)
		// logger.Emergency(msg)
		return errors.New(msg), logger.ALERT
	}

	return nil, logger.DEBUG
}

func (r *ExecStmtResult) Delete() (error, logger.Severity) {
	if nil == r.Pinner {
		msg := "ExecStmtResult_Delete_NULL_PINNER"
		// logger.Error(msg)
		return errors.New(msg), logger.ALERT
	}

	return _ExecStmtResultDelete(r)
}

// @icky
func (r *ExecStmtResult) ContainedDelete(container *container.Container) {
	err, sev := r.Delete()

	if nil != err {
		logger.Log(sev, container.Tag)
	}
}

type Bindable struct {
	Value  string
	IsNull bool
}

func (r *ExecStmtResult) Bind2(values []Bindable) (error, logger.Severity) {
	if nil == r.Pinner {
		msg := "ExecStmtResult_Bind2_NULL_PINNER"
		// logger.Error(msg)
		return errors.New(msg), logger.ALERT
	}

	if 0 == len(values) {
		return nil, logger.DEBUG
	}

	binds := make([]C.wholth_exec_stmt_Bindable, len(values))

	for i, value := range values {
		if value.IsNull {
			binds[i] = C.wholth_exec_stmt_Bindable{C.wholth_StringView{nil, 0}}
		} else {
			binds[i] = C.wholth_exec_stmt_Bindable{toStrView(value.Value)}
		}
	}

	r.BindableCount = uint(len(values))
	r.FirstBindable = (*C.wholth_exec_stmt_Bindable)(unsafe.Pointer(&binds[0]))
	r.Pinner.Pin(r.FirstBindable)

	return nil, logger.DEBUG
}

func (r *ExecStmtResult) Bind(values ...string) (error, logger.Severity) {
	if nil == r.Pinner {
		msg := "ExecStmtResult_Bind_NULL_PINNER"
		// logger.Error(msg)
		return errors.New(msg), logger.ALERT
	}

	// binds := [len(values)]C.wholth_exec_stmt_Bindable{}
	if 0 == len(values) {
		return nil, logger.DEBUG
	}
	binds := make([]C.wholth_exec_stmt_Bindable, len(values))

	for i, value := range values {
		binds[i] = C.wholth_exec_stmt_Bindable{toStrView(value)}
	}

	r.BindableCount = uint(len(values))
	r.FirstBindable = (*C.wholth_exec_stmt_Bindable)(unsafe.Pointer(&binds[0]))
	r.Pinner.Pin(r.FirstBindable)

	return nil, logger.DEBUG
}

// If error was due to validation, then returned `Severity` will
// be equal to `NOTICE`, otherwise it's `ALERT`.
func (r *ExecStmtResult) Fetch(filename string) (error, logger.Severity) {
	if nil == r.Pinner {
		msg := "ExecStmtResult_Fetch_NULL_PINNER"
		// logger.Error(msg)
		return errors.New(msg), logger.ALERT
	}

	args := C.wholth_exec_stmt_Args{
		sql_file:   toStrView(filename),
		binds_size: C.ulonglong(r.BindableCount),
		binds:      r.FirstBindable,
		// TODO skip based on env
		skip_cache: true,
	}
	r.Pinner.Pin(&args)

	werr := C.wholth_exec_stmt(&args, r.Handle)

	if !C.wholth_error_ok(&werr) {
		msg := toStr(werr.message)

		// 19 == SQLITE_CONSTRAINT
		if C.wholth_exec_stmt_Code_BINDABLE_VALIDATION_FAIL == werr.code || 19 == werr.code {
			// logger.Info(msg)
			return errors.New(msg), logger.NOTICE
		} else {
			// logger.Alert(msg)
			return errors.New(msg), logger.ALERT
		}
	}

	return nil, logger.DEBUG
}

func (r *ExecStmtResult) At(row uint, column uint) string {
	return toStr(C.wholth_exec_stmt_Result_at(
		r.Handle,
		C.ulonglong(row),
		C.ulonglong(column),
	))
}

func (r *ExecStmtResult) RowCount() uint64 {
	return uint64(C.wholth_exec_stmt_Result_row_count(r.Handle))
}

func saveSteps(form *PostFoodsForm) error {
	// TODO: handle severity
	result, err, _ := ExecStmtResultNew()

	defer result.Delete()

	if nil != err {
		return err
	}

	re := regexp.MustCompile(`((\d+)(h|ч)){0,1}((\d+)(m|м)){0,1}((\d+)(s|с)){0,1}`)
	res := re.FindAllSubmatch([]byte(form.RecipeStep.Time), 1)

	var seconds int64 = 0
	if nil != res && len(res) == 1 {
		if len(res[0][2]) > 0 {
			hour, err := strconv.ParseInt(string(res[0][2]), 10, 64)
			if nil == err {
				seconds += hour * 60 * 60
			}
		}
		if len(res[0][5]) > 0 {
			minute, err := strconv.ParseInt(string(res[0][5]), 10, 64)
			if nil == err {
				seconds += minute * 60
			}
		}
		if len(res[0][8]) > 0 {
			secs, err := strconv.ParseInt(string(res[0][8]), 10, 64)
			if nil == err {
				seconds += secs
			}
		}
	}
	secondsStr := strconv.FormatInt(seconds, 10)

	err, _ = result.Bind(
		form.Food.Id,
		secondsStr,
		form.Food.Id,
		DEFAULT_LOCALE_ID,
		form.RecipeStep.Description,
	)

	if nil != err {
		return err
	}

	// TODO handle severity
	err, _ = result.Fetch("recipe_step_upsert.sql")

	if nil != err {
		return err
	}

	form.RecipeStep.Id = result.At(0, 0)

	return nil
}

func saveIngredients(buf *C.wholth_Buffer, form *PostFoodsForm) error {
	food := C.wholth_entity_food_init()
	food.id = toStrView(form.Food.Id)

	step := C.wholth_entity_recipe_step_init()
	step.id = toStrView(form.RecipeStep.Id)

	var shouldRecalcNutrients = len(form.Ingredients.Values) > 0

	for i := range form.Ingredients.Values {
		ing := C.wholth_entity_ingredient_init()
		ing.id = toStrView(form.Ingredients.Values[i].Id)
		ing.food_id = toStrView(form.Ingredients.Values[i].FoodId)
		ing.canonical_mass_g = toStrView(form.Ingredients.Values[i].CanonicalMass)

		var err = C.wholth_Error_OK

		if "" != form.Ingredients.Values[i].Id {
			err = C.wholth_em_ingredient_update(&ing, &step, buf)
		} else {
			err = C.wholth_em_ingredient_insert(&ing, &step, buf)
		}

		if !C.wholth_error_ok(&err) {
			form.Ingredients.Values[i].Status.Alias = "error"
			form.Ingredients.Values[i].Status.Message = toStr(err.message)
			shouldRecalcNutrients = false
		} else {
			form.Ingredients.Values[i].Id = toStr(ing.id)
		}
	}

	if shouldRecalcNutrients {
		err := C.wholth_em_food_nutrient_update_important(&food, buf)

		if !C.wholth_error_ok(&err) {
			return errors.New("Ошибка при обновлении основных нутриентов на основе ингредиентов: " + toStr(err.message))
		}
	} else if len(form.Ingredients.Values) > 0 {
		return errors.New("Не удалось сохранить ингридиенты!")
	}

	return nil
}

func SaveFood(form *PostFoodsForm) (string, error) {
	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	err := saveBasics(scratch, form)
	if nil != err {
		return "error", err
	}

	errNutrients := saveNutrients(scratch, form)

	errSteps := saveSteps(form)
	if nil != errSteps {
		form.RecipeStep.Status.Alias = "error"
		form.RecipeStep.Status.Message = errSteps.Error()
		errors.Join(errors.New("Не удалось сохранить рецепт!"), errSteps)
	}

	errIngredients := saveIngredients(scratch, form)

	err = errors.Join(err, errNutrients, errSteps, errIngredients)

	if nil != err {
		return "warning", err
	}

	return "success", nil
}

type DateTime struct {
	Date string
	Time string
}

func (d DateTime) ToWholthFormat() string {
	// TODO fix this. not robust
	if (len(d.Date) + len(d.Time) + 1) == len(DateFormat()) {
		return d.Date + "T" + d.Time
	}

	return d.Date + "T" + d.Time + ":00"
}

func (d DateTime) ToTime() (time.Time, error) {
	// TODO fix this. not robust
	formatted := d.ToWholthFormat()
	return time.Parse(DateFormat(), formatted)
}

func (d *DateTime) UpdateFromTime(t time.Time) {
	formatted := t.Format(DateFormat())
	d.Date = formatted[0:10]
	d.Time = formatted[11:]
}

func DateTimeCreate(date string, time string) DateTime {
	// TODO fix this. not robust
	if len(time) == 0 {
		time = "00:00:00"
	} else if len(time) < len("00:00:00") {
		time = time + ":00"
	}
	// TODO add checks on date field

	return DateTime{Date: date, Time: time}
}
