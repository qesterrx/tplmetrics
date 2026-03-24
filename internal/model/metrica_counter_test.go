package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterUpdateValueAtomic(t *testing.T) {

	var val1 int64 = 1
	var val2 int64 = 2

	mtrk := NewMetricaCounter("a", 1)

	assert.Equal(t, FormatMetricaCounter(val1), mtrk.Value())
	assert.Equal(t, val1, mtrk.SrcValue("new"))
	err := mtrk.UpdateValueAtomic(NewMetricaCounter("a", 1))
	assert.NoError(t, err)
	assert.Equal(t, FormatMetricaCounter(val2), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	err = mtrk.UpdateValueAtomic(NewMetricaCounter("b", 1))
	assert.Error(t, err)
	assert.Equal(t, FormatMetricaCounter(val2), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))

}

func TestCounterUpdateValueConfirm(t *testing.T) {

	var val1 int64 = 1
	var val2 int64 = 2

	mtrk := NewMetricaCounter("a", 1)

	err := mtrk.UpdateValueStart(NewMetricaCounter("a", 1))
	assert.NoError(t, err)
	assert.Equal(t, FormatMetricaCounter(val1), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	assert.Equal(t, val1, mtrk.SrcValue(""))
	mtrk.Confirm()
	assert.Equal(t, FormatMetricaCounter(val2), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	assert.Equal(t, val2, mtrk.SrcValue(""))
}

func TestCounterUpdateValueRestore(t *testing.T) {

	var val1 int64 = 1
	var val2 int64 = 2

	mtrk := NewMetricaCounter("a", 1)

	err := mtrk.UpdateValueStart(NewMetricaCounter("a", 1))
	assert.NoError(t, err)
	assert.Equal(t, FormatMetricaCounter(val1), mtrk.Value())
	assert.Equal(t, val2, mtrk.SrcValue("new"))
	assert.Equal(t, val1, mtrk.SrcValue(""))
	mtrk.Restore()
	assert.Equal(t, FormatMetricaCounter(val1), mtrk.Value())
	assert.Equal(t, val1, mtrk.SrcValue("new"))
	assert.Equal(t, val1, mtrk.SrcValue(""))

}
