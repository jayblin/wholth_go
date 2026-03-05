package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"errors"
	"unsafe"
	"wholth_go/util"
)

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

	C.wholth_pages_recipe_step_locale_id(handle, toStrView(DEFAULT_LOCALE_ID))

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
		// Status: util.Status{
		// 	Alias: "error",
		// 	Message: "AN_ERROR",
		// },
	}

	return result
}
