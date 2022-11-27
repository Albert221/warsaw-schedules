package sql

import (
	"database/sql"
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
	StopComplexID int             `db:"stop_complex_id"`
	StopID        int             `db:"stop_id"`
	Street        string          `db:"street"`
	Direction     string          `db:"direction"`
	Latitude      sql.NullFloat64 `db:"latitude"`
	Longitude     sql.NullFloat64 `db:"longitude"`
	Platform      sql.NullInt16   `db:"platform"`
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
		ST_Y(location) latitude, ST_X(location) longitude, platform
		FROM stops ORDER BY stop_id`)
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
				var location *model.Coordinates
				if stop.Latitude.Valid && stop.Longitude.Valid {
					location = &model.Coordinates{
						Latitude:  stop.Latitude.Float64,
						Longitude: stop.Longitude.Float64,
					}
				}
				var platform *int
				if stop.Platform.Valid {
					tmp := int(stop.Platform.Int16)
					platform = &tmp
				}

				complex.Stops = append(complex.Stops, &model.Stop{
					ID:          fmt.Sprintf("%02d", stop.StopID),
					Street:      stop.Street,
					Direction:   stop.Direction,
					Location:    location,
					Platform:    platform,
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

func (r *SqlStopRepository) SaveStops(stops ...*model.Stop) error {
	valuesStmt := ""
	args := make([]any, 0, len(stops)*6)
	for i, stop := range stops {
		if i > 0 {
			valuesStmt += ", "
		}
		valuesStmt += "(?, ?, ?, ?, ST_GeomFromText(?), ?)"

		point := "null"
		if stop.Location != nil {
			point = fmt.Sprintf("POINT(%f %f)", stop.Location.Longitude, stop.Location.Latitude)
		}

		args = append(args, stop.StopComplex.ID, stop.ID, stop.Street, stop.Direction,
			point, stop.Platform)
	}

	_, err := r.db.Exec("REPLACE INTO stops (stop_complex_id, stop_id, street, direction, location, platform) VALUES "+valuesStmt, args...)

	return err
}
