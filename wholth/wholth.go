package wholth

// #cgo CFLAGS: -I${SRCDIR}/../wholth_lib/include
// #define WHOLTH_GO_INIT
// #include "wholth/wholth.h"
import "C"

import (
	"errors"
	"strings"

	// "fmt"

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
		logger.Emergency(message)
		panic(message)
	}

	C.wholth_app_locale_id(toStrView(DEFAULT_LOCALE_ID))
}

func SetPasswordEncryptionSecret(secret string) {
	C.wholth_app_password_encryption_secret(toStrView(secret))
}


type PostFoodsForm struct {
	Food          Food
	RecipeStep    RecipeStep
	Ingredients   util.PaginatableList[Ingredient]
	Nutrients     util.PaginatableList[FoodNutrient]
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

	if "" != form.Food.Id {
		err := C.wholth_em_food_update(&food, toStrView(DEFAULT_LOCALE_ID), buf)

		if !C.wholth_error_ok(&err) {
			return errors.New("Ошибка сохранения общей инф-ии: " + toStr(err.message))
		}
	} else {
		err := C.wholth_em_food_insert(&food, toStrView(DEFAULT_LOCALE_ID), buf)

		if !C.wholth_error_ok(&err) {
			return errors.New("Ошибка сохранения общей инф-ии: " + toStr(err.message))
		}
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

func saveSteps(buf *C.wholth_Buffer, form *PostFoodsForm) error {
	food := C.wholth_entity_food_init()
	food.id = toStrView(form.Food.Id)

	step := C.wholth_entity_recipe_step_init()
	step.id = toStrView(form.RecipeStep.Id)
	step.description = toStrView(form.RecipeStep.Description)
	step.time = toStrView(form.RecipeStep.Time)

	if "" != form.RecipeStep.Id {
		err := C.wholth_em_recipe_step_update(&step, buf)

		if !C.wholth_error_ok(&err) {
			form.RecipeStep.Status.Alias = "error"
			form.RecipeStep.Status.Message = toStr(err.message)
			return errors.New("Не удалось сохранить рецепт!")
		}
	} else if "" != strings.Trim(form.RecipeStep.Description, " ") {
		err := C.wholth_em_recipe_step_insert(&step, &food, buf)

		if !C.wholth_error_ok(&err) {
			form.RecipeStep.Status.Alias = "error"
			form.RecipeStep.Status.Message = toStr(err.message)
			return errors.New("Не удалось сохранить рецепт!")
		}
	}


	form.RecipeStep.Id = toStr(step.id)

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

		if "" != form.Ingredients.Values[i].Id {
			err := C.wholth_em_ingredient_update(&ing, &step, buf)

			if !C.wholth_error_ok(&err) {
				form.Ingredients.Values[i].Status.Alias = "error"
				form.Ingredients.Values[i].Status.Message = toStr(err.message)
				shouldRecalcNutrients = false
			} else {
				form.Ingredients.Values[i].Id = toStr(ing.id)
			}
		} else {
			err := C.wholth_em_ingredient_insert(&ing, &step, buf)

			if !C.wholth_error_ok(&err) {
				form.Ingredients.Values[i].Status.Alias = "error"
				form.Ingredients.Values[i].Status.Message = toStr(err.message)
				shouldRecalcNutrients = false
			} else {
				form.Ingredients.Values[i].Id = toStr(ing.id)
			}
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

	err = errors.Join(
		err,
		saveNutrients(scratch, form),
		saveSteps(scratch, form),
		saveIngredients(scratch, form),
	)

	if nil != err {
		return "warning", err
	}

	return "success", nil
}
