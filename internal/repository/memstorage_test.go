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
	err = mm.UpdateMetric(model.Counter, "c1", "1")
	require.NoError(t, err)

	err = mm.UpdateMetric(model.Counter, "c1", "1")
	assert.NoError(t, err)
	c1, ok := mm.counter["c1"]
	assert.True(t, ok)
	assert.Equal(t, int64(2), c1)

	//-------Check wrong type
	err = mm.UpdateMetric(model.Counter, "c1", "1.01")
	assert.Error(t, err)

	err = mm.UpdateMetric(model.Counter, "c1", "abs")
	assert.Error(t, err)

	err = mm.UpdateMetric(model.Counter, "c1", "")
	assert.Error(t, err)

	//---Base Gauge
	err = mm.UpdateMetric(model.Gauge, "g1", "1.0001")
	require.NoError(t, err)

	err = mm.UpdateMetric(model.Gauge, "g1", "2.0002")
	assert.NoError(t, err)
	g1, ok := mm.gauge["g1"]
	assert.True(t, ok)
	assert.Equal(t, "2.0002", fmt.Sprintf("%1.4f", g1))

	//-------Check wrong type
	err = mm.UpdateMetric(model.Gauge, "g1", "abs")
	assert.Error(t, err)

	err = mm.UpdateMetric(model.Gauge, "g1", "")
	assert.Error(t, err)

	//---Check have already registr
	err = mm.UpdateMetric(model.Gauge, "c1", "1")
	assert.Error(t, err)

	err = mm.UpdateMetric(model.Counter, "g1", "1")
	assert.Error(t, err)

}

func TestGetMetric(t *testing.T) {

	var err error

	mm := NewMemStorage()
	err = mm.UpdateMetric(model.Counter, "c1", "1")
	require.NoError(t, err)
	err = mm.UpdateMetric(model.Gauge, "g1", "1.0001")
	require.NoError(t, err)

	kind, valC, err := mm.GetMetric("c1")
	assert.NoError(t, err)
	assert.Equal(t, model.Counter, kind)
	assert.Equal(t, "1", valC)

	kind, valG, err := mm.GetMetric("g1")
	assert.NoError(t, err)
	assert.Equal(t, model.Gauge, kind)
	assert.Equal(t, fmt.Sprintf("%f", 1.0001), valG)

	_, _, err = mm.GetMetric("notfound")
	assert.Error(t, err)

}
