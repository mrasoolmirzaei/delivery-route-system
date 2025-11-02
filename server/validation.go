package server

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	maxURLChars = 2048
	maxDstGET   = 80
)

type ValidationError map[string]string

func (e ValidationError) Error() string {
	errors := make([]string, 0)
	for field, message := range e {
		errors = append(errors, fmt.Sprintf("%s: %s", field, message))
	}
	return strings.Join(errors, ", ")
}

func validateGetRoutesRequest(r *http.Request) (*GetRoutesRequest, ValidationError) {
	validationErr := ValidationError{}
	if len(r.URL.String()) > maxURLChars {
		validationErr["url"] = fmt.Sprintf("URL is longer than %d characters", maxURLChars)
	}

	request := &GetRoutesRequest{}
	params := r.URL.Query()

	source := params.Get("src")
	if source == "" {
		validationErr["src"] = "source location is required"
	}

	if source != "" {
		request.Source = Location(source)
		if err := request.Source.Validate(); err != nil {
			validationErr["src"] = err.Error()
		}
	}

	// Get all destination parameters (can be multiple)
	destinations := params["dst"]
	if len(destinations) == 0 {
		validationErr["dst"] = "destination location is required"
	}

	if len(destinations) > maxDstGET {
		validationErr["dst"] = fmt.Sprintf("too many destinations: %d, max is %d", len(destinations), maxDstGET)
	}

	request.Destinations = make([]Location, len(destinations))
	for i, dst := range destinations {
		request.Destinations[i] = Location(dst)
		if err := request.Destinations[i].Validate(); err != nil {
			dstKey := fmt.Sprintf("dst[%d]", i+1)
			validationErr[dstKey] = fmt.Sprintf("destination number %d is invalid: %s", i+1, err.Error())
		}
	}

	if len(validationErr) > 0 {
		return nil, validationErr
	}

	return request, nil
}
