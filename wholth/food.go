package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"errors"
	"unsafe"
	"wholth_go/cache"
)

type Food struct {
	Id          string
	Title       string
	PrepTime    string
	TopNutrient string
}

func (f Food) EntityAlias() string {
	return "food"
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

	werr = C.wholth_pages_food_locale_id(handle, toStrView(DEFAULT_LOCALE_ID))
	if !C.wholth_error_ok(&werr) {
		err = errors.New(toStr(werr.message))
	}

	return FoodPage{Page{handle}}, err
}

func (t *FoodPage) SetTitle(title string) {
	C.wholth_pages_food_title(t.Handle, toStrView(title))
}

// func (t *FoodPage) SetIngredients(titles string) {
// 	C.wholth_pages_food_ingredients(t.Handle, toStrView(titles))
// }

func (t *FoodPage) SetId(id string) {
	C.wholth_pages_food_id(t.Handle, toStrView(id))
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

	cache.Set("g_foods", result.Id, result)

	return result
}
