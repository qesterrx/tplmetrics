package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrica(t *testing.T) {

	/*Этим тестом проверяется только функция NewMetrica */

	mtrk, err := NewMetrica("mtrk_c", string(Counter), "1")
	assert.NoError(t, err)
	assert.Equal(t, "mtrk_c", mtrk.Name())
	assert.Equal(t, Counter, mtrk.Kind())

	mtrk, err = NewMetrica("mtrk_g", string(Gauge), "1")
	assert.NoError(t, err)
	assert.Equal(t, "mtrk_g", mtrk.Name())
	assert.Equal(t, Gauge, mtrk.Kind())

	mtrk, err = NewMetrica("mtrk_g", "UNKNWN", "1")
	assert.Error(t, err)
}
