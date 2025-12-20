package food

import (
	// "fmt"
	"net/http"
	// "net/url"
	"strconv"
	"strings"

	// "net/url"
	// "slices"
	// "strings"
	// "wholth_go/cache"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/route/ingredient"
	"wholth_go/route/nutrient"
	"wholth_go/util"
	"wholth_go/wholth"
)

type GetFoodsPage struct {
	route.HtmlPage
	Values     []wholth.Food
	PostForm   wholth.PostFoodsForm
	Pagination util.Pagination
	FoodQ      string
}

func ListFoods(w http.ResponseWriter, r *http.Request) {
	var limit, limit_err = strconv.Atoi(r.URL.Query().Get("limit"))

	if nil != limit_err || limit < 0 {
		limit = 20
	} else if limit > 50 {
		limit = 50
	}

	var wpage, wpageErr = wholth.FoodPageNew(uint64(limit))

	defer wpage.Close()

	if nil != wpageErr {
		logger.Alert(wpageErr.Error())
		w.Write([]byte(wpageErr.Error()))
		return
	}

	q := r.URL.Query()
	as_subdoc := q.Get("as_subdoc")
	as_suggestion := q.Get("as_suggestion")
	food_q := q.Get("food_q")

	if "" != food_q {
		wpage.SetTitle(strings.ToLower(food_q))
	}

	page_number, page_number_err := strconv.Atoi(r.URL.Query().Get("page_number"))

	if nil == page_number_err && (page_number-1) >= 0 {
		wpage.SkipTo(page_number - 1)
	}

	wpage.Fetch()

	size := wpage.Size()

	var values = make([]wholth.Food, size)

	for i := range size {
		values[i] = wpage.At(i)
	}

	if "" != as_subdoc {
		page := GetFoodsPage{
			Values:     values,
			Pagination: wpage.Pagination(),
			FoodQ:      food_q,
		}
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/food/get/form.html",
			"templates/utils/paginator.html",
		)
	} else if "" != as_suggestion {
		page := GetFoodsPage{
			Values:     values,
			Pagination: wpage.Pagination(),
		}
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/food/get/suggestion_list.html",
			"templates/utils/paginator.html",
		)
	} else {
		page := GetFoodsPage{
			route.DefaultHtmlPage(r),
			values,
			wholth.PostFoodsFormDefault(),
			wpage.Pagination(),
			"",
		}
		page.Meta.Title = "Пища"
		page.Meta.Description = "Список/добавление/изменение пищи."

		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/index.html",
			"templates/food/get/content.html",
			"templates/food/get/form.html",
			"templates/food/post/form.html",
			"templates/ingredient/suggestion_list.html",
			"templates/nutrient/suggestion_list.html",
			"templates/utils/paginator.html",
		)
	}
}

func FindFood(foodId string) wholth.Food {
	result := wholth.Food{}

	var wpage, wpageErr = wholth.FoodPageNew(1)

	defer wpage.Close()

	if nil != wpageErr {
		return result
	}

	wpage.SetId(foodId)

	wpage.Fetch()

	size := wpage.Size()

	if 0 == size {
		// w.WriteHeader(http.StatusNotFound)
		return result
	}

	return wpage.At(0)
}

func FindRecipeStep(foodId string) wholth.RecipeStep {
	result := wholth.RecipeStep{}

	var wpage, wpageErr = wholth.RecipeStepNew()

	defer wpage.Close()

	if nil != wpageErr {
		return result
	}

	wpage.SetId(foodId)

	wpage.Fetch()

	size := wpage.Size()

	if 0 == size {
		// w.WriteHeader(http.StatusNotFound)
		return result
	}

	return wpage.Get()
}

// todo add pagination
// todo query
func FindIngredients(foodId string) []wholth.Ingredient {
	var wpage, wpageErr = wholth.IngredientPageNew(50)

	defer wpage.Close()

	if nil != wpageErr {
		return make([]wholth.Ingredient, 0)
	}

	wpage.SetFoodId(foodId)

	wpage.Fetch()

	size := wpage.Size()

	result := make([]wholth.Ingredient, size)

	for i := range size {
		result[i] = wpage.At(i)
	}

	return result
}

// todo add pagination
// todo query
func FindFoodNutrients(foodId string) []wholth.FoodNutrient {
	var wpage, wpageErr = wholth.FoodNutrientPageNew(50)

	defer wpage.Close()

	if nil != wpageErr {
		return make([]wholth.FoodNutrient, 0)
	}

	wpage.SetFoodId(foodId)

	wpage.Fetch()

	size := wpage.Size()

	result := make([]wholth.FoodNutrient, size)

	for i := range size {
		nut := wpage.At(i)
		nut.Checked = false
		result[i] = nut
	}

	return result
}

func PostFoodsFormFromDb(foodId string) (wholth.PostFoodsForm, int) {
	form := wholth.PostFoodsFormDefault()

	food := FindFood(foodId)

	if "" == food.Id {
		return form, http.StatusNotFound
	}

	form.Food = food
	form.RecipeStep = FindRecipeStep(food.Id)
	form.Ingredients = FindIngredients(food.Id)
	form.Nutrients = FindFoodNutrients(food.Id)

	return form, http.StatusOK
}

func GetFoodById(w http.ResponseWriter, r *http.Request) {
	foodId := r.PathValue("id")
	form, status := PostFoodsFormFromDb(foodId)

	if http.StatusOK != status {
		w.WriteHeader(status)
		return
	}

	page := GetFoodsPage{
		HtmlPage: route.DefaultHtmlPage(r),
		PostForm: form,
	}
	page.Meta.Title = form.Food.Title
	// page.Meta.Description = "Форма алол прикол карбидол"

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/index.html",
		"templates/food/_id/get/content.html",
		"templates/food/post/form.html",
		"templates/ingredient/suggestion_list.html",
		"templates/nutrient/suggestion_list.html",
	)
}

type PostFoodsPage struct {
	route.HtmlPage
	PostForm wholth.PostFoodsForm
}

func PostFoodsFormFromRequest(r *http.Request) wholth.PostFoodsForm {
	r.ParseForm()
	return wholth.PostFoodsForm{
		Food: wholth.Food{
			Id:    r.PostForm.Get("food_id"),
			Title: r.PostForm.Get("food_title"),
		},
		RecipeStep: wholth.RecipeStep{
			Id:          r.PostForm.Get("recipe_step_id"),
			Time:        r.PostForm.Get("recipe_step_time"),
			Description: r.PostForm.Get("recipe_step_description"),
		},
		ResultStatus:  "",
		ResultMessage: "",
		Ingredients:   ingredient.IngredientsFromRequest(r.PostForm),
		Nutrients:     nutrient.FoodNutrientsFromRequest(r.PostForm),
	}
}

func PostFood(w http.ResponseWriter, r *http.Request) {
	page := PostFoodsPage{
		route.DefaultHtmlPage(r),
		PostFoodsFormFromRequest(r)}

	status, err := wholth.SaveFood(&page.PostForm)

	if nil != err {
		page.PostForm.ResultStatus = status
		page.PostForm.ResultMessage = err.Error()
	} else {
		formEnriched, _ := PostFoodsFormFromDb(page.PostForm.Food.Id)

		formEnriched.ResultStatus = status
		formEnriched.ResultMessage = "Успешно сохранено!"

		page.PostForm = formEnriched
	}

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/food/post/form.html",
		"templates/ingredient/suggestion_list.html",
		"templates/nutrient/suggestion_list.html",
	)
}
