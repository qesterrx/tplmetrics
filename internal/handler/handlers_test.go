package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/service"
	"github.com/stretchr/testify/assert"
)

// Это надо будет заменить на mock или придумать что то более красивое
func GetRouterForTest() (*service.TCLService, chi.Router) {

	cfg := config.ConfigServer{
		RestoreFromFileStorage: false,
		StorageMode:            config.MetricaStorageModeSync,
	}
	tcl, _ := service.NewTCLService(&cfg)
	hc := NewHandlerContainer(tcl, "")

	return tcl, hc.GetRouter()

}

func TestGetAllCurrentMetricsHandler(t *testing.T) {

	_, router := GetRouterForTest()

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

func TestGetMetricaHandler(t *testing.T) {

	tcl, router := GetRouterForTest()

	tcl.UpdateMetrica(model.NewMetricaCounter("c1", 5))
	tcl.UpdateMetrica(model.NewMetricaGauge("g1", 5.05005))

	valC, _ := tcl.GetMetrica("c1", string(model.Counter))
	valG, _ := tcl.GetMetrica("g1", string(model.Gauge))

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
			body:       valC.Value(),
		},
		{
			name:       "Get exists gauge",
			url:        "/value/gauge/g1",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			body:       valG.Value(),
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
			statusCode: http.StatusNotFound,
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

func TestUpdateMetricaHandler(t *testing.T) {

	_, router := GetRouterForTest()

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
			name:        "WrongMethod",
			url:         "/update/gauge/g1/1",
			method:      http.MethodGet,
			contentType: "text/plain",
			statusCode:  http.StatusMethodNotAllowed,
		},
		{
			name:        "WrongKind",
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

func TestUpdateMetricaJSONHandler(t *testing.T) {

	_, router := GetRouterForTest()

	tests := []struct {
		name        string
		url         string
		metricaJSON string
		method      string
		contentType string
		statusCode  int
	}{
		{
			name:        "UpdateCounter",
			url:         "/update/",
			metricaJSON: `{"id":"c1","type":"counter","delta":1}`,
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "UpdateGauge",
			url:         "/update/",
			metricaJSON: `{"id":"g1","type":"gauge","value":1.001}`,
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "WrongMethod",
			url:         "/update/",
			metricaJSON: `{"id":"c1","type":"counter","delta":1}`,
			method:      http.MethodGet,
			contentType: "application/json",
			statusCode:  http.StatusMethodNotAllowed,
		},
		{
			name:        "WrongContentType",
			url:         "/update/",
			metricaJSON: `{"id":"c1","type":"counter","delta":1}`,
			method:      http.MethodPost,
			contentType: "text/plain",
			statusCode:  http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, test.url, strings.NewReader(test.metricaJSON))
			r.Header.Set("Content-Type", test.contentType)

			router.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code, fmt.Sprintf("StatusCode in response different: want %d got %d", test.statusCode, w.Code))
			if test.statusCode == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestGetMetricaJSONHandler(t *testing.T) {

	tcl, router := GetRouterForTest()

	tcl.UpdateMetrica(model.NewMetricaCounter("c1", 5))
	tcl.UpdateMetrica(model.NewMetricaGauge("g1", 5.05005))

	reqCounter := `{"id":"c1","type":"counter"}`
	reqGauge := `{"id":"g1","type":"gauge"}`

	resCounter := `{"id":"c1","type":"counter","delta":5}`
	resGauge := `{"id":"g1","type":"gauge","value":5.05005}`

	tests := []struct {
		name       string
		url        string
		req        string
		res        string
		method     string
		statusCode int
	}{
		{
			name:       "Get exists counter",
			url:        "/value/",
			req:        reqCounter,
			res:        resCounter,
			method:     http.MethodPost,
			statusCode: http.StatusOK,
		},
		{
			name:       "Get exists gauge",
			url:        "/value/",
			req:        reqGauge,
			res:        resGauge,
			method:     http.MethodPost,
			statusCode: http.StatusOK,
		},
		{
			name:       "Not exists value",
			url:        "/value/",
			req:        `{"id":"unknwn","type":"counter"}`,
			method:     http.MethodPost,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Not exists kind",
			url:        "/value/",
			req:        `{"id":"c1","type":"unknwn"}`,
			method:     http.MethodPost,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Wrong method",
			url:        "/value/",
			method:     http.MethodGet,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, test.url, strings.NewReader(test.req))
			r.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code, fmt.Sprintf("StatusCode in response different: want %d got %d", test.statusCode, w.Code))

			if w.Code == http.StatusOK {

				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				resp := w.Result()
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)

				assert.NoError(t, err)
				assert.Equal(t, test.res, string(body), fmt.Sprintf("Value in response different: want %s geot %s", test.res, string(body)))
			}
		})

	}
}
