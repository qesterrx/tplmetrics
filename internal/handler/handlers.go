package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

func (hc *HandlerContainer) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Log.Info().Msg("PingDBHandler клиент обратился с ошибочным методом в запросе")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	err := hc.tcl.Check(ctx)
	if err != nil {
		logger.Log.Error().Msg("PingDBHandler не удалось выполнить Ping DB")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (hc *HandlerContainer) GetAllCurrentMetricsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		logger.Log.Info().Msg("GetAllCurrentMetricsHandler клиент обратился с ошибочным методом в запросе")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	metrics := hc.tcl.GetAllMetrics(ctx)

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
}

// Получение метрики по URI
func (hc *HandlerContainer) GetMetricaHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		logger.Log.Info().Msg("GetMetricaHandler клиент обратился с ошибочным методом в запросе")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	kind := strings.TrimSpace(chi.URLParam(r, "kind"))
	name := strings.TrimSpace(chi.URLParam(r, "name"))
	mtrk, err := hc.tcl.GetMetrica(ctx, name, kind)

	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("GetMetricaHandler ошибка при получении метрики %s, %s: %s", kind, name, err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	io.WriteString(w, mtrk.Value())

}

// Обновление значения метрики через URI
func (hc *HandlerContainer) UpdateMetricaHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		logger.Log.Info().Msg("UpdateMetricaHandler клиент обратился с ошибочным методом в запросе")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	kind := strings.TrimSpace(chi.URLParam(r, "kind"))
	name := strings.TrimSpace(chi.URLParam(r, "name"))
	value := strings.TrimSpace(chi.URLParam(r, "value"))

	mtrk, err := model.NewMetrica(name, kind, value)
	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaHandler ошибка создания метрики %s, %s, %s : %s", name, kind, value, err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hc.tcl.UpdateMetrica(ctx, mtrk)
	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaHandler ошибка при обновлении метрики %s, %s, %s : %s", name, kind, value, err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// Обновление значения метрики через json
func (hc *HandlerContainer) UpdateMetricaJSONHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		logger.Log.Info().Msg("UpdateMetricaJSONHandler клиент обратился с ошибочным методом в запросе")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" || r.ContentLength == 0 {
		logger.Log.Info().Msg("UpdateMetricaJSONHandler клиент обратился с ошибочным Content-Type в запросе")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

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

	err = hc.tcl.UpdateMetrica(ctx, mtrk)
	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricaJSONHandler ошибка при обновлении метрики %s : %s", mtrk, err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("{}"))

}

// Получение значения метрики через json
func (hc *HandlerContainer) GetMetricaJSONHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		logger.Log.Info().Msg("GetMetricaJSONHandler клиент обратился с ошибочным методом в запросе")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" || r.ContentLength == 0 {
		logger.Log.Info().Msg("GetMetricaJSONHandler клиент обратился с ошибочным Content-Type в запросе")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	mtrkJSON := model.MetricaJSONAdapter{}
	err := json.NewDecoder(r.Body).Decode(&mtrkJSON)
	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("GetMetricaJSONHandler ошибка разбора JSON: %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mtrk, err := hc.tcl.GetMetrica(ctx, mtrkJSON.Name, mtrkJSON.Kind)

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

}

// Обновление метрик массивом
func (hc *HandlerContainer) UpdateMetricsJSONHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		logger.Log.Info().Msg("UpdateMetricsJSONHandler клиент обратился с ошибочным методом в запросе")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" || r.ContentLength == 0 {
		logger.Log.Info().Msg("UpdateMetricsJSONHandler клиент обратился с ошибочным Content-Type в запросе")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	mtrksJSON := []model.MetricaJSONAdapter{}
	err := json.NewDecoder(r.Body).Decode(&mtrksJSON)
	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricsJSONHandler ошибка разбора JSON: %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mtrks := []model.Metrica{}
	for _, mtrkJSON := range mtrksJSON {
		mtrk, err := mtrkJSON.Metrica()
		if err != nil {
			logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricsJSONHandler ошибка преобразования метрики %v : %s", mtrkJSON, err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mtrks = append(mtrks, mtrk)
	}

	err = hc.tcl.UpdateMetricaBatch(ctx, mtrks)
	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("UpdateMetricsJSONHandler ошибка при обновлении метрик: %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("{}"))

}
