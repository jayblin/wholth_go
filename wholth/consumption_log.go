package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"unsafe"
	"wholth_go/container"
	"wholth_go/util"
)

type ConsumptionLog struct {
	// util.Toggleable
	// util.ToggleableTrait
	Id             string
	FoodId         string
	Mass           string
	NutrientAmount float64
	ConsumedAt     string
	FoodTitle      string
}

func (p ConsumptionLog) EntityAlias() string {
	return "consumption_log"
}

func (p *ConsumptionLog) Toggleable_Checked() bool {
	return false
}

func (p *ConsumptionLog) Toggleable_Name() string {
	return "consumption_log"
}

func (p *ConsumptionLog) Toggleable_Value() string {
	return p.Id
}

type ConsumptionLogPage struct {
	Page
}

func (t *ConsumptionLogPage) SetUserId(userId string) error {
	werr := C.wholth_pages_consumption_log_user_id(t.Handle, toStrView(userId))

	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return err
}

func (t *ConsumptionLogPage) SetPeriod(from time.Time, to time.Time) error {
	werr := C.wholth_pages_consumption_log_period(
		t.Handle,
		toStrView(from.Format(DateFormat())),
		toStrView(to.Format(DateFormat())))

	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return err
}

func (t *ConsumptionLogPage) At(i uint64) ConsumptionLog {
	ptr := &C.wholth_ConsumptionLog{}

	err := C.wholth_pages_at(
		t.Handle,
		(unsafe.Pointer)(ptr),
		C.uint64_t(i))

	if !C.wholth_error_ok(&err) {
		return ConsumptionLog{}
	}

	val := *ptr

	topNut, topNutErr := strconv.ParseFloat(toStr(val.nutrient_amount), 64)

	if nil != topNutErr {
		topNut = 0.0
	}

	result := ConsumptionLog{
		Id:             toStr(val.id),
		FoodId:         toStr(val.food_id),
		Mass:           toStr(val.mass),
		NutrientAmount: topNut,
		ConsumedAt:     toStr(val.consumed_at),
		FoodTitle:      toStr(val.food_title),
	}

	return result
}

func ConsumptionLogPageNew(perPage uint64) (ConsumptionLogPage, error) {
	var handle *C.wholth_Page = nil
	werr := C.wholth_pages_consumption_log(&handle, C.uint64_t(perPage))
	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return ConsumptionLogPage{Page{handle}}, err
}

type ConsumptionLogPostForm struct {
	util.Toggleable
	Id         string
	FoodId     string
	FoodTitle  string
	Mass       string
	ConsumedAt string
	// todo use util struct
	ResultStatus  string
	ResultMessage string
}

func SaveConsumptionLog(form *ConsumptionLogPostForm, userId string) (string, error) {
	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	wuser := C.wholth_entity_user_init()
	wuser.id = toStrView(userId)

	wlog := C.wholth_entity_consumption_log_init()
	wlog.food_id = toStrView(form.FoodId)
	wlog.consumed_at = toStrView(form.ConsumedAt)
	wlog.mass = toStrView(form.Mass)

	var err = C.wholth_Error_OK

	if "" != form.Id {
		wlog.id = toStrView(form.Id)
		err = C.wholth_em_consumption_log_update(&wlog, &wuser, scratch)
	} else {
		err = C.wholth_em_consumption_log_insert(&wlog, &wuser, scratch)
	}

	if !C.wholth_error_ok(&err) {
		return "error", errors.New("Ошибка сохранения лога: " + toStr(err.message))
	}

	form.Id = toStr(wlog.id)

	return "success", nil
}

func UpdateConsumptionLog(r *http.Request, id string, mass string, userId string) error {
	result, err, sev := ExecStmtResultNew()
	defer result.ContainedDelete(container.Instance(r))

	if nil != err {
		container.Log(r, sev, "[UpdateConsumptionLog][0]", err)
		return err
	}

	err, sev = result.Bind(id, userId, mass)

	if nil != err {
		container.Log(r, sev, "[UpdateConsumptionLog][1]", err)
		return err
	}

	err, sev = result.Fetch("consumption_log_update.sql")

	if nil != err {
		container.Log(r, sev, "[UpdateConsumptionLog][2]", err)
		return err
	}

	return nil
}

func DeleteConsumptionLog(r *http.Request, id string, userId string) error {
	result, err, sev := ExecStmtResultNew()
	defer result.ContainedDelete(container.Instance(r))

	if nil != err {
		container.Log(r, sev, "[DeleteConsumptionLog][0]", err)
		return err
	}

	err, sev = result.Bind(id, userId)

	if nil != err {
		container.Log(r, sev, "[DeleteConsumptionLog][1]", err)
		return err
	}

	err, sev = result.Fetch("consumption_log_delete.sql")

	if nil != err {
		container.Log(r, sev, "[DeleteConsumptionLog][2]", err)
		return err
	}

	return nil
}
