package nutrient

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

func FoodNutrientsFromRequest(query url.Values) []wholth.FoodNutrient {
	var values []wholth.FoodNutrient

	nuts, nuts_ok := query["nutrient"]
	nut_values := util.ArrayFilter(
		query["nutrient_value"],
		func(s string) bool {
			return "" != s
		},
	)

	if !nuts_ok || len(nuts) <= 0 {
		return values
	}

	nut_values_len := len(nut_values)

	for idx, nut_id := range nuts {
		if "" == nut_id {
			continue
		}

		var nut_val = ""
		if nut_values_len > idx && "" != nut_values[idx] {
			nut_val = nut_values[idx]
		}

		cached, ok := cache.Get("grp-nutrient", nut_id)

		if !ok || nil == cached {
			value := wholth.FoodNutrient{
				Nutrient: wholth.Nutrient{
					Id:    nut_id,
					Title: "",
					Unit:  "",
				},
				Value:   nut_val,
				Checked: true,
			}
			values = append(values, value)
			continue
		}

		t, ok := cached.(wholth.Nutrient)

		if ok {
			// val := append(t[:3], nut_val)
			// values = append(values, ([4]string)(val))
			val := wholth.FoodNutrient{
				Nutrient: t,
				Value:    nut_val,
			}

			values = append(values, val)
		}
	}

	return values
}

func ListNutrients(w http.ResponseWriter, r *http.Request) {
	// sess, _ := session.Get(r)
	query := r.URL.Query()
	titles_raw := query.Get("nutrient_q")

	if "" == titles_raw {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var values = FoodNutrientsFromRequest(query)

	page, pageErr := wholth.NutrientPageNew(10)

	defer page.Close()

	if nil != pageErr {
		logger.Alert(pageErr.Error())
		w.Write([]byte(pageErr.Error()))
		return
	}

	// todo optimize - should be one call of function inseated of 10 or less.
	titles := strings.SplitN(titles_raw, ",", 10)

	for i := range titles {
		page.SetTitle(titles[i])

		page.Fetch()

		size := page.Size()

		for j := range size {
			nutrient := page.At(j)

			k := slices.IndexFunc(values, func(f wholth.FoodNutrient) bool {
				return f.Id == nutrient.Id
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

			cache.Set("grp-nutrient", nutrient.Id, nutrient)
		}
	}

	route.RenderHtmlTemplates(
		w,
		r,
		values,
		"templates/nutrient/suggestion_list.html",
	)
}
