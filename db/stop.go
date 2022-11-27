package db

import "warsaw-schedules.dev/model"

type StopRepository interface {
	FindAll() ([]*model.StopComplex, error)
	SaveStopComplexes(...*model.StopComplex) error

	SaveCities(...*model.City) error
}
