package server

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name:     "single error",
			err:      ValidationError{"src": "source location is required"},
			expected: "src: source location is required",
		},
		{
			name:     "multiple errors",
			err:      ValidationError{"src": "source location is required", "dst": "destination location is required"},
			expected: "dst: destination location is required, src: source location is required",
		},
		{
			name:     "empty error",
			err:      ValidationError{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if len(tt.err) == 0 {
				assert.Equal(t, tt.expected, result)
			} else {
				for field, message := range tt.err {
					assert.Contains(t, result, field+": "+message)
				}
			}
		})
	}
}

func TestValidateGetRoutesRequest(t *testing.T) {
	tests := []struct {
		name            string
		queryParams     map[string][]string
		wantErr         bool
		wantErrFields   []string
		validateRequest func(*testing.T, *GetRoutesRequest)
	}{
		{
			name: "valid request with single destination",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr: false,
			validateRequest: func(t *testing.T, req *GetRoutesRequest) {
				require.NotNil(t, req)
				assert.Equal(t, Location("12.3456,78.9101"), req.Source)
				assert.Len(t, req.Destinations, 1)
				assert.Equal(t, Location("13.1234,79.9101"), req.Destinations[0])
			},
		},
		{
			name: "valid request with multiple destinations",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"13.1234,79.9101", "14.5678,80.1234"},
			},
			wantErr: false,
			validateRequest: func(t *testing.T, req *GetRoutesRequest) {
				require.NotNil(t, req)
				assert.Equal(t, Location("12.3456,78.9101"), req.Source)
				assert.Len(t, req.Destinations, 2)
				assert.Equal(t, Location("13.1234,79.9101"), req.Destinations[0])
				assert.Equal(t, Location("14.5678,80.1234"), req.Destinations[1])
			},
		},
		{
			name: "missing source",
			queryParams: map[string][]string{
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "missing destinations",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst"},
		},
		{
			name:          "missing source and destinations",
			queryParams:   map[string][]string{},
			wantErr:       true,
			wantErrFields: []string{"src", "dst"},
		},
		{
			name: "invalid source format - not comma separated",
			queryParams: map[string][]string{
				"src": {"12.3456"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid source format - too many parts",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101,90.1234"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid source - invalid latitude",
			queryParams: map[string][]string{
				"src": {"invalid,78.9101"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid source - invalid longitude",
			queryParams: map[string][]string{
				"src": {"12.3456,invalid"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid source - latitude out of range (too high)",
			queryParams: map[string][]string{
				"src": {"90.0,78.9101"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid source - latitude out of range (too low)",
			queryParams: map[string][]string{
				"src": {"-90.0,78.9101"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid source - longitude out of range (too high)",
			queryParams: map[string][]string{
				"src": {"12.3456,180.0"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid source - longitude out of range (too low)",
			queryParams: map[string][]string{
				"src": {"12.3456,-180.0"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"src"},
		},
		{
			name: "invalid destination format",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"13.1234"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst[1]"},
		},
		{
			name: "invalid destination - invalid latitude",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"invalid,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst[1]"},
		},
		{
			name: "invalid destination - invalid longitude",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"13.1234,invalid"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst[1]"},
		},
		{
			name: "invalid destination - latitude out of range",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"90.0,79.9101"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst[1]"},
		},
		{
			name: "invalid destination - longitude out of range",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"13.1234,180.0"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst[1]"},
		},
		{
			name: "multiple invalid destinations",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"invalid1,79.9101", "13.1234,invalid2"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst[1]", "dst[2]"},
		},
		{
			name: "one valid and one invalid destination",
			queryParams: map[string][]string{
				"src": {"12.3456,78.9101"},
				"dst": {"13.1234,79.9101", "invalid,90.0"},
			},
			wantErr:       true,
			wantErrFields: []string{"dst[2]"},
		},
		{
			name: "too many destinations",
			queryParams: func() map[string][]string {
				params := map[string][]string{
					"src": {"12.3456,78.9101"},
					"dst": make([]string, maxDstGET+1),
				}
				for i := 0; i <= maxDstGET; i++ {
					params["dst"][i] = "13.1234,79.9101"
				}
				return params
			}(),
			wantErr:       true,
			wantErrFields: []string{"dst"},
		},
		{
			name: "URL too long",
			queryParams: func() map[string][]string {
				params := map[string][]string{
					"src": {"12.3456,78.9101"},
					"dst": make([]string, 0),
				}
				numDests := (maxURLChars / 25) + 10
				for i := 0; i < numDests; i++ {
					params["dst"] = append(params["dst"], "13.1234,79.9101")
				}
				return params
			}(),
			wantErr:       true,
			wantErrFields: []string{"url"},
		},
		{
			name: "valid edge case - latitude at boundary (just below 90)",
			queryParams: map[string][]string{
				"src": {"89.999999,78.9101"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr: false,
			validateRequest: func(t *testing.T, req *GetRoutesRequest) {
				require.NotNil(t, req)
				assert.Equal(t, Location("89.999999,78.9101"), req.Source)
			},
		},
		{
			name: "valid edge case - longitude at boundary (just below 180)",
			queryParams: map[string][]string{
				"src": {"12.3456,179.999999"},
				"dst": {"13.1234,79.9101"},
			},
			wantErr: false,
			validateRequest: func(t *testing.T, req *GetRoutesRequest) {
				require.NotNil(t, req)
				assert.Equal(t, Location("12.3456,179.999999"), req.Source)
			},
		},
		{
			name: "valid edge case - negative coordinates",
			queryParams: map[string][]string{
				"src": {"-12.3456,-78.9101"},
				"dst": {"-13.1234,-79.9101"},
			},
			wantErr: false,
			validateRequest: func(t *testing.T, req *GetRoutesRequest) {
				require.NotNil(t, req)
				assert.Equal(t, Location("-12.3456,-78.9101"), req.Source)
				assert.Equal(t, Location("-13.1234,-79.9101"), req.Destinations[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := url.URL{
				Path:     "/routes",
				RawQuery: buildQuery(tt.queryParams),
			}
			urlString := u.String()

			req, err := http.NewRequest("GET", "http://example.com"+urlString, nil)
			require.NoError(t, err)

			request, validationErr := validateGetRoutesRequest(req)

			if tt.wantErr {
				require.NotNil(t, validationErr, "expected validation error but got none")
				assert.NotNil(t, validationErr, "validation error should not be nil")
				for _, field := range tt.wantErrFields {
					assert.Contains(t, validationErr, field, "validation error should contain field: %s", field)
				}
				assert.Nil(t, request, "request should be nil when validation fails")
			} else {
				assert.Nil(t, validationErr, "expected no validation error but got: %v", validationErr)
				require.NotNil(t, request, "request should not be nil when validation succeeds")
				if tt.validateRequest != nil {
					tt.validateRequest(t, request)
				}
			}
		})
	}
}

func buildQuery(params map[string][]string) string {
	if len(params) == 0 {
		return ""
	}

	values := url.Values{}
	for key, vals := range params {
		for _, val := range vals {
			values.Add(key, val)
		}
	}
	return values.Encode()
}
