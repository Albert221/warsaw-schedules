package web

import (
	"html/template"
	"log"
	"net/http"

	"warsaw-schedules.dev/db"
	"warsaw-schedules.dev/model"
)

type StopController struct {
	stopRepo db.StopRepository
}

func NewStopController(stopRepo db.StopRepository) *StopController {
	return &StopController{
		stopRepo: stopRepo,
	}
}

func (c *StopController) StopsList() http.Handler {
	tpl := template.Must(template.ParseFiles("web/templates/stops.gohtml"))

	type vm struct {
		StopComplexes []*model.StopComplex
		Rows          int
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stopComplexes, err := c.stopRepo.FindAll()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows := len(stopComplexes)
		for _, stopComplex := range stopComplexes {
			rows += len(stopComplex.Stops)
		}

		vm := vm{
			StopComplexes: stopComplexes,
			Rows:          rows,
		}
		tpl.Execute(w, vm)
	})
}
