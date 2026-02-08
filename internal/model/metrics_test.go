package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetKindValue(t *testing.T) {

	var err error

	_, err = GetKindValue(string(Gauge))
	assert.NoError(t, err)

	_, err = GetKindValue(string(Counter))
	assert.NoError(t, err)

	_, err = GetKindValue("unknown")
	assert.Error(t, err)

}
