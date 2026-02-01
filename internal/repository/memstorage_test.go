package repository

import (
	"fmt"
	"testing"

	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetric(t *testing.T) {

	var err error

	mm := NewMemStorage()

	//---Base Counter
	err = mm.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Counter, Value: 1})
	require.NoError(t, err)

	err = mm.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Counter, Value: 1})
	assert.NoError(t, err)
	c1, ok := mm.counter["c1"]
	assert.True(t, ok)
	assert.Equal(t, int64(2), c1)

	//-------Check wrong type
	err = mm.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Counter, Value: "1.01"})
	assert.Error(t, err)

	err = mm.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Counter, Value: "abs"})
	assert.Error(t, err)

	err = mm.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Counter, Value: ""})
	assert.Error(t, err)

	//---Base Gauge
	err = mm.UpdateMetric(&model.Metrica{Name: "g1", Kind: model.Gauge, Value: 1.0001})
	require.NoError(t, err)

	err = mm.UpdateMetric(&model.Metrica{Name: "g1", Kind: model.Gauge, Value: 2.0002})
	assert.NoError(t, err)
	g1, ok := mm.gauge["g1"]
	assert.True(t, ok)
	assert.Equal(t, "2.0002", fmt.Sprintf("%1.4f", g1))

	//-------Check wrong type
	err = mm.UpdateMetric(&model.Metrica{Name: "g1", Kind: model.Gauge, Value: "abs"})
	assert.Error(t, err)

	err = mm.UpdateMetric(&model.Metrica{Name: "g1", Kind: model.Gauge, Value: ""})
	assert.Error(t, err)

	//---Check have already registr
	err = mm.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Gauge, Value: "1"})
	assert.Error(t, err)

	err = mm.UpdateMetric(&model.Metrica{Name: "g1", Kind: model.Counter, Value: "1"})
	assert.Error(t, err)

}

func TestGetMetric(t *testing.T) {

	var err error

	mm := NewMemStorage()
	err = mm.UpdateMetric(&model.Metrica{Name: "c1", Kind: model.Counter, Value: 1})
	require.NoError(t, err)
	err = mm.UpdateMetric(&model.Metrica{Name: "g1", Kind: model.Gauge, Value: 1.0001})
	require.NoError(t, err)

	metric, err := mm.GetMetric("c1")
	assert.NoError(t, err)
	assert.Equal(t, model.Counter, metric.Kind)
	assert.Equal(t, int64(1), metric.Value)

	metric, err = mm.GetMetric("g1")
	assert.NoError(t, err)
	assert.Equal(t, model.Gauge, metric.Kind)
	assert.Equal(t, float64(1.0001), metric.Value)

	_, err = mm.GetMetric("notfound")
	assert.Error(t, err)

}
