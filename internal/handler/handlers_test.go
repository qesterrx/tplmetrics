package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qesterrx/tplmetrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetric(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/update/{kind}/{name}/{value}", UpdateMetric(repository.NewMemStorage()))

	tests := []struct {
		name        string
		url         string
		method      string
		contentType string
		statusCode  int
	}{
		{
			name:        "UpdateCounter",
			url:         "/update/counter/c1/56",
			method:      http.MethodPost,
			contentType: "text/plain",
			statusCode:  http.StatusOK,
		},
		{
			name:        "UpdateGauge",
			url:         "/update/gauge/g1/1",
			method:      http.MethodPost,
			contentType: "text/plain",
			statusCode:  http.StatusOK,
		},
		{
			name:        "WrongContentType",
			url:         "/update/gauge/g1/1",
			method:      http.MethodPost,
			contentType: "json/application",
			statusCode:  http.StatusBadRequest,
		},
		{
			name:        "WrongMethod",
			url:         "/update/gauge/g1/1",
			method:      http.MethodGet,
			contentType: "text/plain",
			statusCode:  http.StatusBadRequest,
		},
		{
			name:        "WrongKind", //А не нужно ли выносить эту логику в memtorage?
			url:         "/update/unknwn/g1/1",
			method:      http.MethodPost,
			contentType: "text/plain",
			statusCode:  http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, test.url, nil)
			r.Header.Set("Content-Type", test.contentType)

			mux.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code, fmt.Sprintf("StatusCode in response different: want %d got %d", test.statusCode, w.Code))
		})
	}
}
