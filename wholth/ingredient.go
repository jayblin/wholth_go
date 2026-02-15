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

type Ingredient struct {
	util.Status
	Id              string
	FoodId          string
	Title           string
	TopNutrient     string
	PrepTime        string
	CanonicalMass   string
	IngredientsMass string
	Checked         bool
}

func (e Ingredient) EntityAlias() string {
	return "ingredient"
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

	werr = C.wholth_pages_ingredient_locale_id(handle, toStrView(DEFAULT_LOCALE_ID))
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

	ing := *ptr

	result := Ingredient{
		Id:              toStr(ing.id),
		FoodId:          toStr(ing.food_id),
		Title:           toStr(ing.food_title),
		TopNutrient:     "",
		PrepTime:        "",
		CanonicalMass:   toStr(ing.canonical_mass_g),
		IngredientsMass: "",
		// IngredientsMass: toStr(ing.ingredients_mass_g),
	}

	if !cache.Has("g_foods", result.FoodId) {
		cache.Set(
			"g_foods",
			result.FoodId,
			Food{
				Id:          result.FoodId,
				Title:       result.Title,
				PrepTime:    "",
				TopNutrient: "",
			})
	}

	return result
}
