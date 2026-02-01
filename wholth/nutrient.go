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

	werr = C.wholth_pages_nutrient_locale_id(handle, toStrView(DEFAULT_LOCALE_ID))
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

	result := Nutrient{
		Id:    toStr(nut.id),
		Title: toStr(nut.title),
		Unit:  toStr(nut.unit),
	}

	cache.Set("g_nutrients", result.Id, result)

	return result
}
