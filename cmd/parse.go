package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"warsaw-schedules.dev/db"
	"warsaw-schedules.dev/model"
	"warsaw-schedules.dev/parser"
)

var parseCmd = &cobra.Command{
	Use:   "parse file",
	Short: "Parses the given file and populates the database",
	Args:  cobra.ExactArgs(1),
	RunE:  runParse,
}

func runParse(cmd *cobra.Command, args []string) error {
	path := args[0]

	stopRepo := cmd.Context().Value(stopRepoKey).(db.StopRepository)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", path)
		}
		return err
	}
	defer file.Close()

	p := parser.NewParser(file)
	p.OnStopComplexesParsed = func(stopComplexes []parser.StopComplex) error {
		cities := make(map[string]*model.City)
		modelStopComplexes := make([]*model.StopComplex, 0, len(stopComplexes))
		for _, stopComplex := range stopComplexes {
			if _, ok := cities[stopComplex.CityID]; !ok {
				cities[stopComplex.CityID] = &model.City{
					ID:   stopComplex.CityID,
					Name: stopComplex.CityName,
				}
			}

			modelStopComplexes = append(modelStopComplexes, &model.StopComplex{
				ID:   stopComplex.ID,
				Name: stopComplex.Name,
				City: cities[stopComplex.CityID],
			})
		}

		err := stopRepo.SaveCities(maps.Values(cities)...)
		if err != nil {
			return err
		}
		return stopRepo.SaveStopComplexes(modelStopComplexes...)
	}

	p.OnStopsParsed = func(stops []parser.Stop) error {
		modelStops := make([]*model.Stop, 0, len(stops))
		for _, stop := range stops {
			var location *model.Coordinates
			if stop.Latitude != nil && stop.Longitude != nil {
				location = &model.Coordinates{
					Latitude:  *stop.Latitude,
					Longitude: *stop.Longitude,
				}
			}

			modelStops = append(modelStops, &model.Stop{
				ID:          stop.StopID,
				Street:      stop.Street,
				Direction:   stop.Direction,
				Location:    location,
				Platform:    stop.Platform,
				StopComplex: &model.StopComplex{ID: stop.StopComplexID},
			})
		}
		return stopRepo.SaveStops(modelStops...)
	}

	before := time.Now()
	err = p.Parse()
	if err == nil {
		fmt.Printf("file parsed, %v elapsed\n", time.Since(before))
	}
	return err
}
