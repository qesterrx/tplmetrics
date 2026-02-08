package repository

import (
	"testing"

	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetrica(t *testing.T) {

	var err error

	mm := NewMemStorage()

	//---Base Counter
	err = mm.UpdateMetrica(model.NewMetricaCounter("c1", 1))
	require.NoError(t, err)

	err = mm.UpdateMetrica(model.NewMetricaCounter("c1", 1))
	assert.NoError(t, err)
	c1, ok := mm.storage["c1"]
	assert.True(t, ok)
	assert.Equal(t, model.FormatMetricaCounter(int64(2)), c1.Value())

	//---Base Gauge
	err = mm.UpdateMetrica(model.NewMetricaGauge("g1", 1.0001))
	require.NoError(t, err)

	err = mm.UpdateMetrica(model.NewMetricaGauge("g1", 2.2222))
	assert.NoError(t, err)
	g1, ok := mm.storage["g1"]
	assert.True(t, ok)
	assert.Equal(t, model.FormatMetricaGauge(float64(2.2222)), g1.Value())

	//---Check have already registr
	err = mm.UpdateMetrica(model.NewMetricaGauge("c1", 1.0001))
	assert.Error(t, err)

	err = mm.UpdateMetrica(model.NewMetricaCounter("g1", 1))
	assert.Error(t, err)

}

func TestGetMetric(t *testing.T) {

	var err error

	counterValue := int64(1)
	counterGauge := float64(1.1111)

	mm := NewMemStorage()
	err = mm.UpdateMetrica(model.NewMetricaCounter("c1", counterValue))
	require.NoError(t, err)
	err = mm.UpdateMetrica(model.NewMetricaGauge("g1", counterGauge))
	require.NoError(t, err)

	mtrk, err := mm.Metrica("c1")
	assert.NoError(t, err)
	assert.Equal(t, model.Counter, mtrk.Kind())
	assert.Equal(t, model.FormatMetricaCounter(counterValue), mtrk.Value())

	mtrk, err = mm.Metrica("g1")
	assert.NoError(t, err)
	assert.Equal(t, model.Gauge, mtrk.Kind())
	assert.Equal(t, model.FormatMetricaGauge(counterGauge), mtrk.Value())

	_, err = mm.Metrica("notfound")
	assert.Error(t, err)

}
