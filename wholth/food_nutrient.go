package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"errors"
	"unsafe"
	"wholth_go/cache"
	"wholth_go/util"
)

type FoodNutrient struct {
	Nutrient
	util.Status
	Value   string
	Checked bool
}

func (e FoodNutrient) EntityAlias() string {
	// todo rename to food_nutrient
	return "nutrient"
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

	werr = C.wholth_pages_food_nutrient_locale_id(handle, toStrView(DEFAULT_LOCALE_ID))
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

	nutrient := Nutrient{
		Id:    toStr(nut.id),
		Title: toStr(nut.title),
		Unit:  toStr(nut.unit),
	}
	cache.Set("g_nutrients", nutrient.Id, nutrient)

	return FoodNutrient{
		Nutrient: nutrient,
		Value:    toStr(nut.value),

		// Status: util.Status{
		// 	Alias: "warning",
		// 	Message: "AN_ERROR",
		// },
	}
}
