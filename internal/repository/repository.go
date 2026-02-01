package repository

import (
	"github.com/qesterrx/tplmetrics/internal/model"
)

type Repository interface {
	UpdateMetric(kind model.KindValue, name string, value string) error
	GetMetric(name string) (model.KindValue, string, error)
	Show()
}
