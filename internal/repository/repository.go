package repository

import (
	"github.com/qesterrx/tplmetrics/internal/model"
)

type Repository interface {
	UpdateMetric(metrica *model.Metrica) error
	GetMetric(name string) (*model.Metrica, error)
	GetAllMetric() *[]model.Metrica
	ShowAllMetric()
}
