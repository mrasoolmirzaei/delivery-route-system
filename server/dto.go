package server

type Location string

func (l Location) String() string {
	return string(l)
}

type GetRoutesRequest struct {
	Source Location
	Destinations []Location
}

type GetRoutesResponse struct {
	Source Location	`json:"source"`
	Routes []*Route `json:"routes"`
}

type Route struct {
	Destination Location `json:"destination"`
	Distance    float64    `json:"distance"`
	Duration    float64    `json:"duration"`
}