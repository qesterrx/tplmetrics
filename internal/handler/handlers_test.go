package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetric(t *testing.T) {

	storage := repository.NewMemStorage()

	router := GetRouter(storage)

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
		/*	{
			name:        "WrongContentType",
			url:         "/update/gauge/g1/1",
			method:      http.MethodPost,
			contentType: "json/application",
			statusCode:  http.StatusBadRequest,
		},*/
		{
			name:        "WrongMethod",
			url:         "/update/gauge/g1/1",
			method:      http.MethodGet,
			contentType: "text/plain",
			statusCode:  http.StatusMethodNotAllowed,
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

			router.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code, fmt.Sprintf("StatusCode in response different: want %d got %d", test.statusCode, w.Code))
		})
	}
}

func TestGetMetric(t *testing.T) {

	storage := repository.NewMemStorage()

	storage.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Counter, Value: "5"})
	storage.UpdateMetric(&model.Metrica{Name: "g1", Kind: model.Gauge, Value: "5.05005"})

	val_c, _ := storage.GetMetric("c1")
	val_g, _ := storage.GetMetric("g1")

	router := GetRouter(storage)

	tests := []struct {
		name       string
		url        string
		method     string
		body       string
		statusCode int
	}{
		{
			name:       "Get exists counter",
			url:        "/value/counter/c1",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			body:       val_c.GetMetricaValue(),
		},
		{
			name:       "Get exists gauge",
			url:        "/value/gauge/g1",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			body:       val_g.GetMetricaValue(),
		},
		{
			name:       "Not exists value",
			url:        "/value/gauge/qwe1",
			method:     http.MethodGet,
			statusCode: http.StatusNotFound,
			body:       "",
		},
		{
			name:       "Not exists kind",
			url:        "/value/wtf/g1",
			method:     http.MethodGet,
			statusCode: http.StatusBadRequest,
			body:       "",
		},
		{
			name:       "Wrong method",
			url:        "/value/wtf/g1",
			method:     http.MethodPost,
			statusCode: http.StatusMethodNotAllowed,
			body:       "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, test.url, nil)

			router.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code, fmt.Sprintf("StatusCode in response different: want %d got %d", test.statusCode, w.Code))

			if test.body != "" {
				resp := w.Result()
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)

				assert.NoError(t, err)
				assert.Equal(t, test.body, string(body), fmt.Sprintf("Value in response different: want %s geot %s", test.body, string(body)))
			}
		})

	}
}

func TestGetAllCurrentMetric(t *testing.T) {

	storage := repository.NewMemStorage()

	router := GetRouter(storage)

	tests := []struct {
		name       string
		method     string
		body       string
		statusCode int
	}{
		{
			name:       "Correct method",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
		},
		{
			name:       "Wrong method",
			method:     http.MethodPost,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, "/", nil)

			router.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code, fmt.Sprintf("StatusCode in response different: want %d got %d", test.statusCode, w.Code))

		})

	}
}
