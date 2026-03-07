package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/qesterrx/tplmetrics/internal/compression"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/repository"
)

func GetRouter(storage repository.MetricaStorage) chi.Router {
	r := chi.NewRouter()

	r.Use(logger.LoggingMiddleware)
	r.Use(compression.GzipCompressMiddleware)

	r.Get(`/value/{kind}/{name}`, GetMetricaHandler(storage))
	r.Get(`/`, GetAllCurrentMetricsHandler(storage))
	r.Get(`/ping`, PingDBHandler(storage))
	r.Post(`/update/{kind}/{name}/{value}`, UpdateMetricaHandler(storage))
	r.Post(`/update/`, UpdateMetricaJSONHandler(storage))
	r.Post(`/value/`, GetMetricaJSONHandler(storage))

	return r
}

func PingDBHandler(storage repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.Log.Info().Msg("PingDBHandler клиент обратился с ошибочным методом в запросе")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if PGStorage, ok := storage.(*repository.PGStorage); ok {
			err := PGStorage.PingDB()
			if err != nil {
				logger.Log.Error().Msg("PingDBHandler не удалось выполнить Ping DB")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		} else {
			logger.Log.Info().Msg("PingDBHandler сервер запущен без работы с БД, метод не поддерживается")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

func GetAllCurrentMetricsHandler(storage repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.Log.Info().Msg("GetAllCurrentMetricsHandler клиент обратился с ошибочным методом в запросе")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metrics := storage.AllMetrics()

		tmpl := `
		<html>
    		<head><title>Current metrics</title></head>
			<body>
				<h2>Current metrics</h2>
				<ul>
                {{range .Metric}}
                	<li>{{.Name}} ({{.Kind}}): {{.Value}}</li>
                {{end}}
            	</ul>
			</body>
		</html>`

		data := struct {
			Metric []model.Metrica
		}{
			Metric: metrics,
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		t, _ := template.New("home").Parse(tmpl)
		t.Execute(w, data)
	})
}

// Получение метрики по URI
func GetMetricaHandler(ms repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			logger.Log.Info().Msg("GetMetricaHandler клиент обратился с ошибочным методом в запросе")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		kind := strings.TrimSpace(chi.URLParam(r, "kind"))
		name := strings.TrimSpace(chi.URLParam(r, "name"))
		mtrk, err := ms.Metrica(name, kind)

		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("GetMetricaHandler ошибка при получении метрики %s, %s: %s", kind, name, err.Error()))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		io.WriteString(w, mtrk.Value())
		w.WriteHeader(http.StatusOK)

	})
}

// Обновление значения метрики через URI
func UpdateMetricaHandler(ms repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			logger.Log.Info().Msg("UpdateMetricaHandler клиент обратился с ошибочным методом в запросе")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		kind := strings.TrimSpace(chi.URLParam(r, "kind"))
		name := strings.TrimSpace(chi.URLParam(r, "name"))
		value := strings.TrimSpace(chi.URLParam(r, "value"))

		mtrk, err := model.NewMetrica(name, kind, value)
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaHandler ошибка создания метрики %s, %s, %s : %s", name, kind, value, err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ms.UpdateMetrica(mtrk)
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaHandler ошибка при обновлении метрики %s, %s, %s : %s", name, kind, value, err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)

	})
}

// Обновление значения метрики через json
func UpdateMetricaJSONHandler(ms repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" || r.ContentLength == 0 {
			logger.Log.Info().Msg("UpdateMetricaJSONHandler клиент обратился с ошибочным методом в запросе")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mtrkJSON := model.MetricaJSONAdapter{}
		err := json.NewDecoder(r.Body).Decode(&mtrkJSON)
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaJSONHandler ошибка разбора JSON: %s", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mtrk, err := mtrkJSON.Metrica()
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaJSONHandler ошибка преобразования метрики %v : %s", mtrkJSON, err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ms.UpdateMetrica(mtrk)
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaJSONHandler ошибка при обновлении метрики %s : %s", mtrk, err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("{}"))

	})
}

// Получение значения метрики через json
func GetMetricaJSONHandler(ms repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" || r.ContentLength == 0 {
			logger.Log.Info().Msg("GetMetricaJSONHandler клиент обратился с ошибочным методом в запросе")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mtrkJSON := model.MetricaJSONAdapter{}
		err := json.NewDecoder(r.Body).Decode(&mtrkJSON)
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("GetMetricaJSONHandler ошибка разбора JSON: %s", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mtrk, err := ms.Metrica(mtrkJSON.Name, mtrkJSON.Kind)

		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("GetMetricaJSONHandler ошибка при получении метрики %s, %s: %s", mtrkJSON.Name, mtrkJSON.Kind, err.Error()))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		body, err := json.Marshal(&mtrk)
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("GetMetricaJSONHandler сериализации метрики %s: %s", mtrk, err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(body)

	})
}
