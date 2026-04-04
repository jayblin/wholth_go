package food

import (
	// "fmt"
	"net/http"
	// "net/url"
	"strconv"

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

type ListFoodsPage struct {
	route.HtmlPage
	util.PaginatableList[wholth.Food]
	PostForm wholth.PostFoodsForm
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
	food_q := util.ModifyQueryForSearch(q.Get("q"))

	if "" != food_q {
		wpage.SetTitle(food_q)
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
	page := ListFoodsPage{}
	page.Q = food_q
	page.Values = values
	page.Pagination = wpage.Pagination()

	if "" != q.Get("as_radio") {
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/food/get/as_subdoc/index.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",

			"templates/food/get/as_subdoc/content.html",
			"templates/food/get/as_radio/suggestion.html",
		)
	} else if "" != q.Get("as_subdoc") {
		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/food/get/as_subdoc/index.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",

			"templates/food/get/as_subdoc/content.html",
			"templates/food/get/suggestion.html",
		)
	} else {
		page.HtmlPage = route.DefaultHtmlPage(r)
		page.PostForm = wholth.PostFoodsFormDefault()
		page.Meta.Title = "Пища"
		page.Meta.Description = "Список/добавление/изменение пищи."

		route.RenderHtmlTemplates(
			w,
			r,
			page,
			"templates/index.html",

			"templates/utils/search.html",
			"templates/utils/paginator.html",

			"templates/food/get/content.html",
			"templates/food/get/suggestion.html",
			"templates/food/post/form.html",

			"templates/ingredient/list_item.html",
			"templates/nutrient/list_item.html",
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

func findRecipeStep(foodId string) wholth.RecipeStep {
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
func findIngredients(foodId string) util.PaginatableList[wholth.Ingredient] {
	var wpage, wpageErr = wholth.IngredientPageNew(50)

	defer wpage.Close()

	result := util.PaginatableList[wholth.Ingredient]{}

	if nil != wpageErr {
		result.Values = make([]wholth.Ingredient, 0)
		return result
	}

	wpage.SetFoodId(foodId)

	wpage.Fetch()

	size := wpage.Size()

	result.Values = make([]wholth.Ingredient, size)
	result.Pagination = wpage.Pagination()

	for i := range size {
		result.Values[i] = wpage.At(i)
	}

	return result
}

// todo add pagination
// todo query
func findFoodNutrients(foodId string) util.PaginatableList[wholth.FoodNutrient] {
	var wpage, wpageErr = wholth.FoodNutrientPageNew(50)

	defer wpage.Close()

	result := util.PaginatableList[wholth.FoodNutrient]{}

	if nil != wpageErr {
		result.Values = make([]wholth.FoodNutrient, 0)
		return result
	}

	wpage.SetFoodId(foodId)

	wpage.Fetch()

	size := wpage.Size()

	result.Values = make([]wholth.FoodNutrient, size)
	result.Pagination = wpage.Pagination()

	for i := range size {
		nut := wpage.At(i)
		nut.Checked = false
		result.Values[i] = nut
	}

	return result
}

func fetchPostFoodsFormFromDb(foodId string) (wholth.PostFoodsForm, int) {
	form := wholth.PostFoodsFormDefault()

	food := FindFood(foodId)

	if "" == food.Id {
		return form, http.StatusNotFound
	}

	form.Food = food
	form.RecipeStep = findRecipeStep(food.Id)
	form.Ingredients = findIngredients(food.Id)
	form.Nutrients = findFoodNutrients(food.Id)

	return form, http.StatusOK
}

func GetFoodById(w http.ResponseWriter, r *http.Request) {
	foodId := r.PathValue("id")
	form, status := fetchPostFoodsFormFromDb(foodId)

	if http.StatusOK != status {
		w.WriteHeader(status)
		return
	}

	page := ListFoodsPage{
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

		"templates/utils/search.html",
		"templates/utils/paginator.html",

		"templates/food/_id/get/content.html",
		"templates/food/post/form.html",

		"templates/ingredient/list_item.html",
		"templates/nutrient/list_item.html",
	)
}

func populateTopNutrients(form *wholth.PostFoodsForm) {
	result, err := wholth.ExecStmtResultNew()
	defer result.Delete()

	if nil != err {
		logger.Error("populateTopNutrients_STMT_RES_INIT: " + err.Error())
		return
	}

	err = result.Bind(wholth.DEFAULT_LOCALE_ID)

	if nil != err {
		logger.Error("populateTopNutrients_STMT_RES_BIND: " + err.Error())
		return
	}

	err = result.Fetch("nutrient_top_select.sql")

	if nil != err {
		logger.Error("populateTopNutrients_STMT_RES_FETCH: " + err.Error())
		return
	}

	sz := result.RowCount()

	if sz == 0 {
		return
	}

	form.Nutrients.Count = sz
	form.Nutrients.PageCurrent = 0
	form.Nutrients.PageMax = 0
	form.Nutrients.Values = make([]wholth.FoodNutrient, sz)
	for i := range uint(sz) {
		form.Nutrients.Values[i] = wholth.FoodNutrient{
			Nutrient: wholth.Nutrient{
				Id:    result.At(i, 0),
				Title: result.At(i, 1),
				Unit:  result.At(i, 2),
			},
			Checked: false,
		}
	}
}

func GetRecipeForm(w http.ResponseWriter, r *http.Request) {
	form := wholth.PostFoodsForm{}

	populateTopNutrients(&form)

	page := ListFoodsPage{
		HtmlPage: route.DefaultHtmlPage(r),
		PostForm: form,
	}
	page.Meta.Title = form.Food.Title

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/index.html",

		"templates/utils/search.html",
		"templates/utils/paginator.html",

		"templates/food/_id/get/content.html",
		"templates/food/post/form.html",

		"templates/ingredient/list_item.html",
		"templates/nutrient/list_item.html",
	)
}

type PostFoodsPage struct {
	route.HtmlPage
	PostForm wholth.PostFoodsForm
}

func PostFoodsFormFromRequest(r *http.Request) wholth.PostFoodsForm {
	r.ParseForm()
	result := wholth.PostFoodsForm{
		Food: wholth.Food{
			Id:    r.PostForm.Get("food_id"),
			Title: r.PostForm.Get("food_title"),
		},
		RecipeStep: wholth.RecipeStep{
			Id:          r.PostForm.Get("recipe_step_id"),
			Time:        r.PostForm.Get("recipe_step_time"),
			Description: r.PostForm.Get("recipe_step_description"),
		},
		// Status: util.Status{},

		// ResultStatus:  "",
		// ResultMessage: "",
	}
	result.Ingredients.Values = ingredient.IngredientsFromRequest(r.PostForm)
	result.Nutrients.Values = nutrient.FoodNutrientsFromRequest(r.PostForm)

	return result
}

func PostFood(w http.ResponseWriter, r *http.Request) {

	page := PostFoodsPage{
		route.DefaultHtmlPage(r),
		PostFoodsFormFromRequest(r)}

	status, err := wholth.SaveFood(&page.PostForm)

	if nil != err {
		page.PostForm.Status.Alias = status
		page.PostForm.Status.Message = err.Error()
		// page.PostForm.ResultStatus = status
		// page.PostForm.ResultMessage = err.Error()
	} else {
		formEnriched, _ := fetchPostFoodsFormFromDb(page.PostForm.Food.Id)

		// formEnriched.ResultStatus = status
		// formEnriched.ResultMessage = "Успешно сохранено!"
		page.PostForm.Status.Alias = status
		page.PostForm.Status.Message = "Успешно сохранено!"

		page.PostForm = formEnriched

		// http.Redirect(
		// 	w,
		// 	r,
		// 	fmt.Sprintf("/food/%s", formEnriched.Food.Id),
		// 	http.StatusSeeOther,
		// )
	}

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/food/post/form.html",

		"templates/utils/search.html",
		"templates/utils/paginator.html",

		"templates/ingredient/list_item.html",
		"templates/nutrient/list_item.html",
	)
}
