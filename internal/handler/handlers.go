package handler

import (
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/repository"
)

func GetRouter(storage repository.MetricaStorage) chi.Router {
	r := chi.NewRouter()

	r.Post(`/update/{kind}/{name}/{value}`, logger.LoggingMiddleware(UpdateMetricaHandler(storage)))
	r.Get(`/value/{kind}/{name}`, logger.LoggingMiddleware(GetMetricaHandler(storage)))
	r.Get(`/`, logger.LoggingMiddleware(GetAllCurrentMetricsHandler(storage)))

	return r
}

func GetAllCurrentMetricsHandler(storage repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			//fmt.Printf("GetMetric, wrong methos, got %s \n", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metrics := storage.AllMetrica()

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

		t, _ := template.New("home").Parse(tmpl)
		t.Execute(w, data)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}

func GetMetricaHandler(ms repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			//fmt.Printf("GetMetric, wrong methos, got %s \n", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		kindSrc := strings.ToLower(chi.URLParam(r, "kind"))
		kind, err := model.GetKindValue(kindSrc)

		if err != nil {
			//fmt.Printf("UpdateMetric, error kind, got %s \n", kindSrc)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		name := strings.ToLower(chi.URLParam(r, "name"))

		mtrk, err := ms.Metrica(name)

		if err != nil {
			//fmt.Printf("GetMetric, error GetMetric,%e \n", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if kind != mtrk.Kind() {
			//fmt.Printf("GetMetric, error kind got %s current %s \n", kind, mtrk.Kind())
			w.WriteHeader(http.StatusNotFound)
			return
		}

		io.WriteString(w, mtrk.Value())
		w.WriteHeader(http.StatusOK)

	})
}

func UpdateMetricaHandler(ms repository.MetricaStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			//fmt.Printf("UpdateMetric, wrong method, got %s \n", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		/*	if r.Header.Get("Content-Type") != "text/plain" {
			fmt.Printf("UpdateMetric, wrong content-type, got %s \n", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}*/

		kindSrc := strings.ToLower(chi.URLParam(r, "kind"))
		kind, err := model.GetKindValue(kindSrc)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		name := strings.ToLower(chi.URLParam(r, "name"))
		value := strings.TrimSpace(chi.URLParam(r, "value"))

		mtrk, err := model.NewMetrica(name, kind, value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = ms.UpdateMetrica(mtrk)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//ms.ShowDebug()

		w.WriteHeader(http.StatusOK)

	})
}
