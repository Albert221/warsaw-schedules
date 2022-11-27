package model

type StopComplex struct {
	ID    string
	Name  string
	City  *City
	Stops []*Stop
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type Stop struct {
	ID          string
	Street      string
	Direction   string
	Location    *Coordinates
	Platform    *int
	StopComplex *StopComplex
}

func (s *Stop) FullID() string {
	return s.StopComplex.ID + s.ID
}
