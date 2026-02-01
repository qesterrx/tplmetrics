package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/repository"
)

func GetRouter(storage repository.Repository) chi.Router {
	r := chi.NewRouter()

	r.Post(`/update/{kind}/{name}/{value}`, UpdateMetric(storage))
	r.Get(`/value/{kind}/{name}`, GetMetric(storage))
	r.Get(`/`, GetAllCurrentMetric(storage))

	return r
}

func GetAllCurrentMetric(storage repository.Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			fmt.Printf("GetMetric, wrong methos, got %s \n", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metrics := storage.GetAllMetric()
		json.NewEncoder(w).Encode(metrics)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}

func GetMetric(storage repository.Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			fmt.Printf("GetMetric, wrong methos, got %s \n", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		kindSrc := strings.ToLower(chi.URLParam(r, "kind"))
		kind, err := model.GetKindValue(kindSrc)

		if err != nil {
			fmt.Printf("UpdateMetric, error kind, got %s \n", kindSrc)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		name := strings.ToLower(chi.URLParam(r, "name"))

		metrica, err := storage.GetMetric(name)

		if err != nil {
			fmt.Printf("GetMetric, error GetMetric,%e \n", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if kind != metrica.Kind {
			fmt.Printf("GetMetric, error kind got %s current %s \n", kind, metrica.Kind)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		io.WriteString(w, metrica.GetMetricaValue())
		w.WriteHeader(http.StatusOK)

	})
}

func UpdateMetric(storage repository.Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			fmt.Printf("UpdateMetric, wrong method, got %s \n", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			fmt.Printf("UpdateMetric, wrong content-type, got %s \n", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		kindSrc := strings.ToLower(chi.URLParam(r, "kind"))
		kind, err := model.GetKindValue(kindSrc)

		if err != nil {
			fmt.Printf("UpdateMetric, error kind, got %s \n", kindSrc)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		name := strings.ToLower(chi.URLParam(r, "name"))
		value := strings.TrimSpace(chi.URLParam(r, "value"))

		//fmt.Println("kind", kind)
		//fmt.Println("name", name)
		//fmt.Println("value", value)

		err = storage.UpdateMetric(&model.Metrica{Name: name, Kind: kind, Value: fmt.Sprintf("%v", value)})

		if err != nil {
			fmt.Printf("UpdateMetric error main %v \n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		storage.ShowAllMetric()

		w.WriteHeader(http.StatusOK)

	})
}
