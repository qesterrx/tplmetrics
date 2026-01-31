package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/qesterrx/tplmetrics/internal/repository"
)

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

		kindSrc := strings.ToLower(r.PathValue("kind"))
		kind, err := repository.GetKindValue(kindSrc)

		if err != nil {
			fmt.Printf("UpdateMetric, error kind, got %s \n", kindSrc)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		name := strings.ToLower(r.PathValue("name"))
		value := strings.TrimSpace(r.PathValue("value"))

		fmt.Println("kind", kind)
		fmt.Println("name", name)
		fmt.Println("value", value)

		err = storage.UpdateMetric(kind, name, value)

		if err != nil {
			fmt.Printf("UpdateMetric error main %v \n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		storage.Show()

		w.WriteHeader(http.StatusOK)

	})
}
