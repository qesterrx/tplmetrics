package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGaugeUpdateValueAtomic(t *testing.T) {

	var val1 float64
	val1 = 1.111111111111111111
	var val2 float64
	val2 = 2.222222222222222222

	mtrk := NewMetricaGauge("a", val1)

	assert.Equal(t, FormatMetricaGauge(val1), mtrk.Value())
	assert.Equal(t, val1, mtrk.SrcValue("new"))
	err := mtrk.UpdateValueAtomic(NewMetricaGauge("a", val2))
	assert.NoError(t, err)
	assert.Equal(t, FormatMetricaGauge(val2), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	err = mtrk.UpdateValueAtomic(NewMetricaGauge("b", val1))
	assert.Error(t, err)
	assert.Equal(t, FormatMetricaGauge(val2), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))

}

func TestGaugeUpdateValueConfirm(t *testing.T) {

	var val1 float64
	val1 = 1.111111111111111111
	var val2 float64
	val2 = 2.222222222222222222

	mtrk := NewMetricaGauge("a", val1)

	err := mtrk.UpdateValueStart(NewMetricaGauge("a", val2))
	assert.NoError(t, err)
	assert.Equal(t, FormatMetricaGauge(val1), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	assert.Equal(t, val1, mtrk.SrcValue(""))
	mtrk.Confirm()
	assert.Equal(t, FormatMetricaGauge(val2), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	assert.Equal(t, val2, mtrk.SrcValue(""))
}

func TestGaugeUpdateValueRestore(t *testing.T) {

	var val1 float64
	val1 = 1.111111111111111111
	var val2 float64
	val2 = 2.222222222222222222

	mtrk := NewMetricaGauge("a", val1)

	err := mtrk.UpdateValueStart(NewMetricaGauge("a", val2))
	assert.NoError(t, err)
	assert.Equal(t, FormatMetricaGauge(val1), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	assert.Equal(t, val1, mtrk.SrcValue(""))
	mtrk.Restore()
	assert.Equal(t, FormatMetricaGauge(val1), mtrk.Value())
	assert.Equal(t, val1, mtrk.SrcValue("new"))
	assert.Equal(t, val1, mtrk.SrcValue(""))

}
