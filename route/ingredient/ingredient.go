package ingredient

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"wholth_go/cache"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/util"
	"wholth_go/wholth"
)

func IngredientsFromRequest(query url.Values) []wholth.Ingredient {
	var values []wholth.Ingredient

	ings, ing_ok := query["ingredient"]
	if !ing_ok || len(ings) <= 0 {
		return values
	}

	ing_masses := util.ArrayFilter(
		query["ingredient_mass"],
		func(s string) bool {
			return "" != s
		},
	)
	ing_ids := util.ArrayFilter(
		query["ingredient_id"],
		func(s string) bool {
			return "" != s
		},
	)

	for idx, food_id := range ings {
		if "" == food_id {
			continue
		}

		var mass = ""
		if len(ing_masses) > idx && "" != ing_masses[idx] {
			mass = ing_masses[idx]
		}

		var ing_id = ""
		if len(ing_ids) > idx && "" != ing_ids[idx] {
			ing_id = ing_ids[idx]
		}

		// todo change group
		cached, ok := cache.Get("grp-food", food_id)

		if !ok || nil == cached {
			food_arr := wholth.Ingredient{
				Id:            ing_id,
				FoodId:        food_id,
				Title:         food_id,
				TopNutrient:   "",
				PrepTime:      "",
				CanonicalMass: mass,
			}
			values = append(values, food_arr)
			continue
		}

		t, ok := cached.(wholth.Food)

		if ok {
			// val := append(t[:4], mass)
			// values = append(values, ([5]string)(val))
			val := wholth.Ingredient{
				Id:            "",
				FoodId:        t.Id,
				Title:         t.Title,
				TopNutrient:   t.TopNutrient,
				PrepTime:      t.PrepTime,
				CanonicalMass: mass,
			}
			values = append(values, val)
		}
	}

	return values
}

func ListIngredients(w http.ResponseWriter, r *http.Request) {
	// sess, _ := session.Get(r)
	query := r.URL.Query()
	titles_raw := query.Get("ingredient_q")

	if "" == titles_raw {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	page, pageErr := wholth.FoodPageNew(10)

	defer page.Close()

	if nil != pageErr {
		logger.Alert(pageErr.Error())
		w.Write([]byte(pageErr.Error()))
		return
	}

	// todo optimize - should be one call of function inseated of 10 or less.
	titles := strings.SplitN(titles_raw, ",", 10)

	var values = IngredientsFromRequest(query)

	for i := range titles {
		page.SetTitle(titles[i])

		page.Fetch()

		size := page.Size()

		for j := range size {
			ing := page.At(j)

			k := slices.IndexFunc(values, func(f wholth.Ingredient) bool {
				return f.FoodId == ing.Id
			})

			if -1 != k {
				continue
			}

			values = append(values, wholth.Ingredient{
				Id:            "",
				FoodId:        ing.Id,
				Title:         ing.Title,
				TopNutrient:   ing.TopNutrient,
				PrepTime:      ing.PrepTime,
				CanonicalMass: "",
			})

			// todo move to page.At()
			cache.Set("grp-food", ing.Id, ing)
		}
	}

	// titles := strings.Join(strings.SplitN(titles_raw, ",", 10), ",")
	// C.wholth_pages_food_ingredients(handle, ToStrView(titles))
	//
	// // todo check fo errors
	// C.wholth_pages_fetch(handle)
	//
	// size := C.wholth_pages_food_array_size(handle)
	// var values = make([][4]string, size)
	//
	// for i := C.ulonglong(0); i < size; i++ {
	// 	food := C.wholth_pages_food_array_at(i, handle)
	// 	food_arr := [4]string{
	// 		ToStr(food.title),
	// 		ToStr(food.top_nutrient),
	// 		ToStr(food.preparation_time),
	// 		ToStr(food.id),
	// 	}
	// 	values[i] = food_arr
	// }

	route.RenderHtmlTemplates(
		w,
		r,
		values,
		"templates/ingredient/suggestion_list.html",
	)
}
