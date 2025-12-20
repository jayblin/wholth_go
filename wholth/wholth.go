package wholth

// #cgo CFLAGS: -I${SRCDIR}/../wholth_lib/include
// #include "wholth/wholth.h"
import "C"

import (
	"errors"
	"time"
	// "fmt"
	"unsafe"
	"wholth_go/logger"
	"wholth_go/secret"
	"wholth_go/util"
)

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
		logger.Emergency(message)
		panic(message)
	}

	C.wholth_app_locale_id(toStrView("2"))
}

func SetPasswordEncryptionSecret(secret string) {
	C.wholth_app_password_encryption_secret(toStrView(secret))
}

func UserRegister(username string, password string) (string, error) {
	if !secret.GetAllowRegistration() {
		return "", errors.New("Регистрация в данный момент не разрешена!")
	}

	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	wuser := C.wholth_entity_user_init()
	wuser.name = toStrView(username)
	wuser.locale_id = toStrView("2")
	wpassword := toStrView(password)
	// fmt.Println(username, password)
	werr := C.wholth_em_user_insert(&wuser, wpassword, scratch)

	if !C.wholth_error_ok(&werr) {
		return "", errors.New(toStr(werr.message))
	}

	return toStr(wuser.id), nil
}

func UserExists(username string) (string, error) {
	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	wusername := toStrView(username)
	var wid = C.wholth_StringView{}

	werr := C.wholth_em_user_exists(wusername, &wid, scratch)

	if !C.wholth_error_ok(&werr) {
		return "", errors.New(toStr(werr.message))
	}

	return toStr(wid), nil
}

func UserAuthenticate(username string, password string) (string, error) {
	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	wuser := C.wholth_entity_user_init()
	wuser.name = toStrView(username)
	wpassword := toStrView(password)
	werr := C.wholth_em_user_authenticate(&wuser, wpassword, scratch)

	if !C.wholth_error_ok(&werr) {
		return "", errors.New(toStr(werr.message))
	}

	return toStr(wuser.id), nil
}

type Page struct {
	Handle *C.wholth_Page
}

func (t *Page) Fetch() error {
	err := C.wholth_pages_fetch(t.Handle)

	if !C.wholth_error_ok(&err) {
		return errors.New(toStr(err.message))
	}

	return nil
}

func (t *Page) Size() uint64 {
	return uint64(C.wholth_pages_array_size(t.Handle))
}

func (t *Page) Close() {
	C.wholth_pages_del(t.Handle)
}

func (t *Page) Pagination() util.Pagination {
	return util.Pagination{
		PageCurrent: uint64(C.wholth_pages_current_page_num(t.Handle)) + 1,
		PageMax:     uint64(C.wholth_pages_max(t.Handle)) + 1,
		Count:       uint64(C.wholth_pages_count(t.Handle)),
	}
}

type Food struct {
	Id          string
	Title       string
	PrepTime    string
	TopNutrient string
}

type FoodPage struct {
	Page
}

func FoodPageNew(perPage uint64) (FoodPage, error) {
	var handle *C.wholth_Page = nil

	werr := C.wholth_pages_food(&handle, C.uint64_t(perPage))
	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return FoodPage{Page{handle}}, err
}

func (t *FoodPage) SetTitle(title string) {
	C.wholth_pages_food_title(t.Handle, toStrView(title))
}

func (t *FoodPage) SetId(id string) {
	C.wholth_pages_food_id(t.Handle, toStrView(id))
}

func (t *FoodPage) SkipTo(to int) {
	C.wholth_pages_skip_to(t.Handle, C.uint64_t(to))
}

func (t *FoodPage) At(i uint64) Food {
	ptr := &C.wholth_Food{}

	err := C.wholth_pages_at(
		t.Handle,
		(unsafe.Pointer)(ptr),
		C.uint64_t(i))

	if !C.wholth_error_ok(&err) {
		return Food{}
	}

	val := *ptr

	result := Food{
		Id:          toStr(val.id),
		Title:       toStr(val.title),
		PrepTime:    toStr(val.preparation_time),
		TopNutrient: toStr(val.top_nutrient),
	}

	return result
}

type RecipeStep struct {
	util.Status
	Id          string
	Time        string
	Description string
}

type RecipeStepPage struct {
	Page
}

func RecipeStepNew() (RecipeStepPage, error) {
	var handle *C.wholth_Page = nil
	werr := C.wholth_pages_recipe_step(&handle)
	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return RecipeStepPage{Page{handle}}, err
}

func (t *RecipeStepPage) SetId(id string) {
	C.wholth_pages_recipe_step_recipe_id(t.Handle, toStrView(id))
}

func (t *RecipeStepPage) Get() RecipeStep {
	ptr := &C.wholth_RecipeStep{}

	err := C.wholth_pages_at(
		t.Handle,
		(unsafe.Pointer)(ptr),
		0)

	if !C.wholth_error_ok(&err) {
		return RecipeStep{}
	}

	val := *ptr

	result := RecipeStep{
		Id:          toStr(val.id),
		Time:        toStr(val.time),
		Description: toStr(val.description),
	}

	return result
}

type Ingredient struct {
	util.Status
	Id            string
	FoodId        string
	Title         string
	TopNutrient   string
	PrepTime      string
	CanonicalMass string
}

type IngredientPage struct {
	Page
}

func IngredientPageNew(perPage uint64) (IngredientPage, error) {
	var handle *C.wholth_Page = nil
	werr := C.wholth_pages_ingredient(&handle, C.uint64_t(perPage))
	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return IngredientPage{Page{handle}}, err
}

func (t *IngredientPage) SetFoodId(id string) IngredientPage {
	var handle *C.wholth_Page = nil
	// todo add err check
	C.wholth_pages_ingredient_food_id(t.Handle, toStrView(id))

	return IngredientPage{Page{handle}}
}

// todo extract to template function
func (t *IngredientPage) At(i uint64) Ingredient {
	ptr := &C.wholth_Ingredient{}

	err := C.wholth_pages_at(
		t.Handle,
		unsafe.Pointer(ptr),
		C.uint64_t(i))

	if !C.wholth_error_ok(&err) {
		return Ingredient{}
	}

	food := *ptr

	return Ingredient{
		Id:            toStr(food.id),
		FoodId:        toStr(food.food_id),
		Title:         toStr(food.food_title),
		TopNutrient:   "",
		PrepTime:      "",
		CanonicalMass: toStr(food.canonical_mass_g),
	}
}

type Nutrient struct {
	Id    string
	Title string
	Unit  string
}

type NutrientPage struct {
	Page
}

func NutrientPageNew(perPage uint64) (NutrientPage, error) {
	var handle *C.wholth_Page = nil
	werr := C.wholth_pages_nutrient(&handle, C.uint64_t(perPage))

	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return NutrientPage{Page{handle}}, err
}

func (t *NutrientPage) SetTitle(title string) {
	C.wholth_pages_nutrient_title(t.Handle, toStrView(title))
}

func (t *NutrientPage) At(i uint64) Nutrient {
	ptr := &C.wholth_Nutrient{}

	err := C.wholth_pages_at(
		t.Handle,
		unsafe.Pointer(ptr),
		C.uint64_t(i))

	if !C.wholth_error_ok(&err) {
		return Nutrient{}
	}

	nut := *ptr

	return Nutrient{
		toStr(nut.id),
		toStr(nut.title),
		toStr(nut.unit),
	}
}

type FoodNutrient struct {
	Nutrient
	util.Status
	Value   string
	Checked bool
}

type FoodNutrientPage struct {
	Page
}

func FoodNutrientPageNew(perPage uint64) (FoodNutrientPage, error) {
	var handle *C.wholth_Page = nil
	werr := C.wholth_pages_food_nutrient(&handle, C.uint64_t(perPage))
	var err error = nil

	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return FoodNutrientPage{Page{handle}}, err
}

func (t *FoodNutrientPage) SetFoodId(id string) {
	C.wholth_pages_food_nutrient_food_id(t.Handle, toStrView(id))
}

func (t *FoodNutrientPage) At(i uint64) FoodNutrient {
	ptr := &C.wholth_Nutrient{}

	err := C.wholth_pages_at(
		t.Handle,
		unsafe.Pointer(ptr),
		C.uint64_t(i))

	if !C.wholth_error_ok(&err) {
		return FoodNutrient{}
	}

	nut := *ptr

	return FoodNutrient{
		Nutrient: Nutrient{
			Id:    toStr(nut.id),
			Title: toStr(nut.title),
			Unit:  toStr(nut.unit),
		},
		Value: toStr(nut.value),
	}
}

type ConsumptionLog struct {
	Id         string
	FoodId     string
	Mass       string
	ConsumedAt string
	FoodTitle  string
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

	result := ConsumptionLog{
		Id:         toStr(val.id),
		FoodId:     toStr(val.food_id),
		Mass:       toStr(val.mass),
		ConsumedAt: toStr(val.consumed_at),
		FoodTitle:  toStr(val.food_title),
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
	Id            string
	FoodId        string
	FoodTitle     string
	Mass          string
	ConsumedAt    string
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

type PostFoodsForm struct {
	// Id            string
	// Title         string
	Food         Food
	RecipeStep   RecipeStep
	ResultStatus string
	// RecipeStepId  string
	ResultMessage string
	// PrepTime      string
	Ingredients []Ingredient
	Nutrients   []FoodNutrient
}

func PostFoodsFormDefault() PostFoodsForm {
	return PostFoodsForm{
		// Id:            "",
		// ResultStatus:  "",
		// ResultMessage: "",
		// // RecipeStepId:  "",
		// RecipeStep:    RecipeStep{},
		// // PrepTime:      "",
		// Ingredients:   []Ingredient{},
		// Nutrients:     []FoodNutrient{},
	}
}

func SaveFood(form *PostFoodsForm) (string, error) {
	var scratch *C.wholth_Buffer = nil

	defer C.wholth_buffer_del(scratch)

	C.wholth_buffer_new(&scratch)

	food := C.wholth_entity_food_init()
	food.id = toStrView(form.Food.Id)
	food.title = toStrView(form.Food.Title)

	// if (svs[2].size > 0) {
	// 	food.description = svs[2]
	// }

	var err = C.wholth_Error_OK

	if "" != form.Food.Id {
		err = C.wholth_em_food_update(&food, scratch)
	} else {
		err = C.wholth_em_food_insert(&food, scratch)
	}

	if !C.wholth_error_ok(&err) {
		return "error", errors.New("Ошибка сохранения общей инф-ии: " + toStr(err.message))
	}

	var status = "success"
	// copying food.id from scratch-buffer
	foodId := toStr(food.id)
	food.id = toStrView(foodId)

	for i := range form.Nutrients {
		wnut := C.wholth_entity_nutrient_init()
		wnut.id = toStrView(form.Nutrients[i].Id)
		wnut.value = toStrView(form.Nutrients[i].Value)

		var err = C.wholth_em_food_nutrient_upsert(&food, &wnut, scratch)

		if !C.wholth_error_ok(&err) {
			form.Nutrients[i].Status.Alias = "error"
			form.Nutrients[i].Status.Message = toStr(err.message)
		}
	}

	step := C.wholth_entity_recipe_step_init()
	step.id = toStrView(form.RecipeStep.Id)
	step.description = toStrView(form.RecipeStep.Description)
	step.time = toStrView(form.RecipeStep.Time)

	if "" != form.RecipeStep.Id {
		err = C.wholth_em_recipe_step_update(&step, scratch)
	} else {
		err = C.wholth_em_recipe_step_insert(&step, &food, scratch)
	}

	if !C.wholth_error_ok(&err) {
		form.RecipeStep.Alias = "error"
		form.RecipeStep.Message = toStr(err.message)
		status = "warning"
		return status, errors.New("Ошибка сохранения рецепта")
	}

	for i := range form.Ingredients {
		ing := C.wholth_entity_ingredient_init()
		ing.id = toStrView(form.Ingredients[i].Id)
		ing.food_id = toStrView(form.Ingredients[i].FoodId)
		ing.canonical_mass_g = toStrView(form.Ingredients[i].CanonicalMass)

		if "" != form.Ingredients[i].Id {
			err = C.wholth_em_ingredient_update(&ing, &step, scratch)
		} else {
			err = C.wholth_em_ingredient_insert(&ing, &step, scratch)
		}

		if !C.wholth_error_ok(&err) {
			form.Ingredients[i].Status.Alias = "error"
			form.Ingredients[i].Status.Message = toStr(err.message)
			status = "warning"
		} else {
			form.Ingredients[i].Id = toStr(ing.id)
		}
	}

	err = C.wholth_em_food_nutrient_update_important(&food, scratch)

	if !C.wholth_error_ok(&err) {
		return "warning", errors.New("Ошибка при обнове основных нутриентов: " + toStr(err.message))
	}

	return status, nil
}
