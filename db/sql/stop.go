package sql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"warsaw-schedules.dev/model"
)

type SqlStopRepository struct {
	db *sqlx.DB
}

type cityEntity struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

type stopComplexEntity struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	CityID string `db:"city_id"`
}

type stopEntity struct {
	StopComplexID int     `db:"stop_complex_id"`
	StopID        int     `db:"stop_id"`
	Street        string  `db:"street"`
	Direction     string  `db:"direction"`
	Latitude      float64 `db:"latitude"`
	Longitude     float64 `db:"longitude"`
}

func NewSqlStopRepository(db *sqlx.DB) *SqlStopRepository {
	return &SqlStopRepository{db: db}
}

func (r *SqlStopRepository) FindAll() ([]*model.StopComplex, error) {
	var cities []cityEntity
	err := r.db.Select(&cities, "SELECT * FROM cities")
	if err != nil {
		return nil, err
	}

	citiesMap := make(map[string]*model.City)
	for _, city := range cities {
		citiesMap[city.ID] = &model.City{
			ID:   city.ID,
			Name: city.Name,
		}
	}

	var stopComplexes []stopComplexEntity
	err = r.db.Select(&stopComplexes, "SELECT * FROM stop_complexes ORDER BY name")
	if err != nil {
		return nil, err
	}

	var stops []stopEntity
	err = r.db.Select(&stops, `SELECT stop_complex_id, stop_id, street, direction, 
		ST_Y(location) latitude, ST_X(location) longitude FROM stops ORDER BY stop_id`)
	if err != nil {
		return nil, err
	}

	result := make([]*model.StopComplex, 0, len(stopComplexes))
	for _, stopComplex := range stopComplexes {
		complex := &model.StopComplex{
			ID:    fmt.Sprintf("%04d", stopComplex.ID),
			Name:  stopComplex.Name,
			City:  citiesMap[stopComplex.CityID],
			Stops: make([]*model.Stop, 0),
		}

		for i := 0; i < len(stops); i++ {
			stop := stops[i]
			if stop.StopComplexID == stopComplex.ID {
				complex.Stops = append(complex.Stops, &model.Stop{
					ID:          fmt.Sprintf("%02d", stop.StopID),
					Street:      stop.Street,
					Direction:   stop.Direction,
					Latitude:    stop.Latitude,
					Longitude:   stop.Longitude,
					StopComplex: complex,
				})

				// remove the stop from the slice
				stops = append(stops[:i], stops[i+1:]...)
				i--
			}
		}

		result = append(result, complex)
	}

	return result, nil
}

func (r *SqlStopRepository) SaveStopComplexes(stopComplexes ...*model.StopComplex) error {
	valuesStmt := ""
	args := make([]any, 0, len(stopComplexes)*3)
	for i, stopComplex := range stopComplexes {
		if i > 0 {
			valuesStmt += ", "
		}
		valuesStmt += "(?, ?, ?)"

		args = append(args, stopComplex.ID, stopComplex.Name, stopComplex.City.ID)
	}

	_, err := r.db.Exec("REPLACE INTO stop_complexes (id, name, city_id) VALUES "+valuesStmt, args...)

	return err
}

func (r *SqlStopRepository) SaveCities(cities ...*model.City) error {
	valuesStmt := ""
	args := make([]any, 0, len(cities)*2)
	for i, city := range cities {
		if i > 0 {
			valuesStmt += ", "
		}
		valuesStmt += "(?, ?)"

		args = append(args, city.ID, city.Name)
	}

	_, err := r.db.Exec("REPLACE INTO cities (id, name) VALUES "+valuesStmt, args...)

	return err
}
