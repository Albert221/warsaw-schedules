package model

type StopComplex struct {
	ID    string
	Name  string
	City  *City
	Stops []*Stop
}

type Stop struct {
	ID          string
	Street      string
	Direction   string
	Latitude    float64
	Longitude   float64
	StopComplex *StopComplex
}

func (s *Stop) FullID() string {
	return s.StopComplex.ID + s.ID
}
