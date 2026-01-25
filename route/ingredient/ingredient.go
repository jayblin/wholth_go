package ingredient

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"wholth_go/cache"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/util"

	// "wholth_go/util"
	"wholth_go/wholth"
)

func IngredientsFromRequest(query url.Values) []wholth.Ingredient {
	var values []wholth.Ingredient
	// fmt.Println(query)

	food_ids, ing_ok := query["ingredient"]
	if !ing_ok || len(food_ids) <= 0 {
		return values
	}

	for _, food_id := range food_ids {

		ingredient_id := query[fmt.Sprintf("ingredient_%s_id", food_id)]
		ingredient_mass := query[fmt.Sprintf("ingredient_%s_mass", food_id)]

		if len(ingredient_id) != 1 || len(ingredient_mass) != 1 {
			continue
		}

		id := ingredient_id[0]
		mass := ingredient_mass[0]

		ing := wholth.Ingredient{
			Id:            id,
			FoodId:        food_id,
			Title:         food_id,
			TopNutrient:   "",
			PrepTime:      "",
			CanonicalMass: mass,
			Checked:       true,
		}

		cached, ok := cache.Get("g_foods", food_id)

		if ok && nil != cached {
			food := cached.(wholth.Food)
			ing.Title = food.Title
			ing.TopNutrient = food.TopNutrient
			ing.PrepTime = food.PrepTime
		}

		values = append(values, ing)
		continue
	}

	return values
}

type ListIngredientsPage struct {
	route.HtmlPage
	util.PaginatableList[wholth.Ingredient]
}

func ListIngredients(w http.ResponseWriter, r *http.Request) {
	// sess, _ := session.Get(r)
	query := r.URL.Query()
	titles_raw := query.Get("q")

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

	var values = IngredientsFromRequest(query)

	page.SetTitle(titles_raw)

	page_number, page_number_err := strconv.Atoi(query.Get("page_number"))

	if nil == page_number_err && (page_number-1) >= 0 {
		page.SkipTo(page_number - 1)
	}

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
	}

	htmlPage := ListIngredientsPage{
		HtmlPage: route.DefaultHtmlPage(r),
		PaginatableList: util.PaginatableList[wholth.Ingredient]{
			Values:     values,
			Pagination: page.Pagination(),
			Q:          titles_raw,
		},
	}
	// htmlPage.Meta.Title = "d"

	route.RenderHtmlTemplates(
		w,
		r,
		htmlPage,
		"templates/ingredient/index.html",

		"templates/utils/search.html",
		"templates/utils/paginator.html",

		"templates/ingredient/suggestion.html",
		"templates/ingredient/list_item.html",
	)
}
