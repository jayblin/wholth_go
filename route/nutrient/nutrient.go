package nutrient

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	// "wholth_go/cache"
	"wholth_go/cache"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/util"

	// "wholth_go/util"
	"wholth_go/wholth"
)

func FoodNutrientsFromRequest(query url.Values) []wholth.FoodNutrient {
	var values []wholth.FoodNutrient

	ids, nuts_ok := query["nutrient"]

	if !nuts_ok || len(ids) <= 0 {
		return values
	}

	for _, id := range ids {
		nutrient_data := query[fmt.Sprintf("nutrient_%s", id)]

		if len(nutrient_data) != 1 {
			continue
		}

		value := nutrient_data[0]

		fn := wholth.FoodNutrient{
			Nutrient: wholth.Nutrient{
				Id:    id,
				Title: id,
				Unit:  "попугай",
			},
			Value:   value,
			Checked: true,
		}

		cached, ok := cache.Get("g_nutrients", id)

		if ok && nil != cached {
			nut := cached.(wholth.Nutrient)
			fn.Title = nut.Title
			fn.Unit = nut.Unit
		}

		values = append(values, fn)
	}

	return values
}

type ListFoodNutrientsPage struct {
	route.HtmlPage
	util.PaginatableList[wholth.FoodNutrient]
}

func ListNutrients(w http.ResponseWriter, r *http.Request) {
	// sess, _ := session.Get(r)
	query := r.URL.Query()
	titles_raw := util.ModifyQueryForSearch(query.Get("q"))

	if "" == titles_raw {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Не указан обязательный параметр: 'q'!"))
		return
	}

	var values = FoodNutrientsFromRequest(query)

	page, pageErr := wholth.NutrientPageNew(50)

	defer page.Close()

	if nil != pageErr {
		logger.Alert(pageErr.Error())
		w.Write([]byte(pageErr.Error()))
		return
	}

	page.SetTitle(titles_raw)

	page_number, page_number_err := strconv.Atoi(query.Get("page_number"))

	if nil == page_number_err && (page_number-1) >= 0 {
		page.SkipTo(page_number - 1)
	}

	page.Fetch()

	size := page.Size()

	for j := range size {
		nutrient := page.At(j)

		k := slices.IndexFunc(values, func(fn wholth.FoodNutrient) bool {
			return fn.Id == nutrient.Id
		})

		if -1 != k {
			continue
		}

		values = append(
			values,
			wholth.FoodNutrient{
				Nutrient: nutrient,
				Value:    "",
			},
		)
	}

	htmlPage := ListFoodNutrientsPage{
		HtmlPage: route.DefaultHtmlPage(r),
		PaginatableList: util.PaginatableList[wholth.FoodNutrient]{
			Values:     values,
			Pagination: page.Pagination(),
			Q:          titles_raw,
		},
	}

	route.RenderHtmlTemplates(
		w,
		r,
		htmlPage,
		"templates/nutrient/index.html",

		"templates/utils/search.html",
		"templates/utils/paginator.html",

		"templates/nutrient/suggestion.html",
		"templates/nutrient/list_item.html",
	)
}
