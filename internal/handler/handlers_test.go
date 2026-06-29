package handler

import (
	"context"
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
func GetRouterForTest(t *testing.T) (*service.TCLService, chi.Router) {

	cfg := config.ConfigServer{
		RestoreFromFileStorage: false,
		StorageMode:            config.MetricaStorageModeSync,
	}
	tcl, _ := service.NewTCLService(&cfg)
	hc := NewHandlerContainer(tcl, "", nil, nil)

	return tcl, hc.GetRouter()

}

func TestGetAllCurrentMetricsHandler(t *testing.T) {

	_, router := GetRouterForTest(t)

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

	tcl, router := GetRouterForTest(t)
	ctx := context.Background()

	tcl.UpdateMetrica(ctx, model.NewMetricaCounter("c1", 5))
	tcl.UpdateMetrica(ctx, model.NewMetricaGauge("g1", 5.05005))

	valC, _ := tcl.GetMetrica(ctx, "c1", string(model.Counter))
	valG, _ := tcl.GetMetrica(ctx, "g1", string(model.Gauge))

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

	_, router := GetRouterForTest(t)

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

	_, router := GetRouterForTest(t)

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
		{
			name:        "EmptyContentType",
			url:         "/update/",
			metricaJSON: `{"id":"c1","type":"counter","delta":1}`,
			method:      http.MethodPost,
			contentType: "",
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

	tcl, router := GetRouterForTest(t)
	ctx := context.Background()

	tcl.UpdateMetrica(ctx, model.NewMetricaCounter("c1", 5))
	tcl.UpdateMetrica(ctx, model.NewMetricaGauge("g1", 5.05005))

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
		{
			name:       "Wrong deserialization",
			url:        "/value/",
			req:        `no json`,
			method:     http.MethodPost,
			statusCode: http.StatusBadRequest,
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

func TestPingDBHandler(t *testing.T) {
	_, router := GetRouterForTest(t)

	tests := []struct {
		name       string
		method     string
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
		{
			name:       "Wrong method PUT",
			method:     http.MethodPut,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, "/ping", nil)

			router.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code)
		})
	}
}

func TestUpdateMetricsJSONHandler(t *testing.T) {
	_, router := GetRouterForTest(t)

	tests := []struct {
		name        string
		url         string
		metricsJSON string
		method      string
		contentType string
		statusCode  int
	}{
		{
			name: "Update multiple metrics",
			url:  "/updates/",
			metricsJSON: `[
				{"id":"c1","type":"counter","delta":10},
				{"id":"g1","type":"gauge","value":15.5},
				{"id":"c2","type":"counter","delta":5}
			]`,
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name: "Update single metric in batch",
			url:  "/updates/",
			metricsJSON: `[
				{"id":"c1","type":"counter","delta":1}
			]`,
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name:        "Empty batch",
			url:         "/updates/",
			metricsJSON: `[]`,
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusOK,
		},
		{
			name: "Invalid JSON",
			url:  "/updates/",
			metricsJSON: `[
				{"id":"c1","type":"counter","delta":10},
				{"id":"g1","type":"gauge","value":15.5},
			]`, // Трейлинг запятая - невалидный JSON
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusBadRequest,
		},
		{
			name:        "Wrong method",
			url:         "/updates/",
			metricsJSON: `[{"id":"c1","type":"counter","delta":1}]`,
			method:      http.MethodGet,
			contentType: "application/json",
			statusCode:  http.StatusMethodNotAllowed,
		},
		{
			name:        "Wrong content type",
			url:         "/updates/",
			metricsJSON: `[{"id":"c1","type":"counter","delta":1}]`,
			method:      http.MethodPost,
			contentType: "text/plain",
			statusCode:  http.StatusBadRequest,
		},
		{
			name:        "Empty body",
			url:         "/updates/",
			metricsJSON: "",
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusBadRequest,
		},
		{
			name: "Invalid metric type",
			url:  "/updates/",
			metricsJSON: `[
				{"id":"c1","type":"invalid","delta":10}
			]`,
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusBadRequest,
		},
		{
			name: "Missing required field",
			url:  "/updates/",
			metricsJSON: `[
				{"id":"c1","type":"counter"}
			]`,
			method:      http.MethodPost,
			contentType: "application/json",
			statusCode:  http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var body io.Reader
			if test.metricsJSON != "" {
				body = strings.NewReader(test.metricsJSON)
			} else {
				body = nil
			}
			r := httptest.NewRequest(test.method, test.url, body)
			if test.contentType != "" {
				r.Header.Set("Content-Type", test.contentType)
			}

			router.ServeHTTP(w, r)

			assert.Equal(t, test.statusCode, w.Code)

			if test.statusCode == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				resp := w.Result()
				defer resp.Body.Close()

				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, "{}", string(bodyBytes))
			}
		})
	}
}
