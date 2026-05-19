package repository

import (
	"context"
	"testing"

	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetrica(t *testing.T) {

	var err error
	ctx := context.Background()

	mm := NewMemStorage()

	//---Base Counter
	err = mm.UpdateMetrica(ctx, model.NewMetricaCounter("c1", 1))
	require.NoError(t, err)

	err = mm.UpdateMetrica(ctx, model.NewMetricaCounter("c1", 1))
	assert.NoError(t, err)

	c1, err := mm.GetMetrica(ctx, "c1", string(model.Counter))

	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaCounter(int64(2)), c1.Value())

	//---Base Gauge
	err = mm.UpdateMetrica(ctx, model.NewMetricaGauge("g1", 1.0001))
	require.NoError(t, err)

	err = mm.UpdateMetrica(ctx, model.NewMetricaGauge("g1", 2.2222))
	assert.NoError(t, err)

	g1, err := mm.GetMetrica(ctx, "g1", string(model.Gauge))

	assert.NoError(t, err)
	assert.Equal(t, model.FormatMetricaGauge(float64(2.2222)), g1.Value())

}

func TestGetMetric(t *testing.T) {

	ctx := context.Background()
	var err error

	counterValue := int64(1)
	counterGauge := float64(1.1111)

	mm := NewMemStorage()
	err = mm.UpdateMetrica(ctx, model.NewMetricaCounter("c1", counterValue))
	require.NoError(t, err)
	err = mm.UpdateMetrica(ctx, model.NewMetricaGauge("g1", counterGauge))
	require.NoError(t, err)

	mtrk, err := mm.GetMetrica(ctx, "c1", string(model.Counter))
	assert.NoError(t, err)
	assert.Equal(t, model.Counter, mtrk.Kind())
	assert.Equal(t, model.FormatMetricaCounter(counterValue), mtrk.Value())

	mtrk, err = mm.GetMetrica(ctx, "g1", string(model.Gauge))
	assert.NoError(t, err)
	assert.Equal(t, model.Gauge, mtrk.Kind())
	assert.Equal(t, model.FormatMetricaGauge(counterGauge), mtrk.Value())

	_, err = mm.GetMetrica(ctx, "notfound", "counter")
	assert.Error(t, err)

}
