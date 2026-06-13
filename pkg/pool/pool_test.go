package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тестируемая структура с методом Reset
type ResetStruct struct {
	Value int
	Name  string
}

func (t *ResetStruct) Reset() {
	t.Value = 0
	t.Name = ""
}

func TestPoolResetableStruct(t *testing.T) {

	//Создание пула с корректной функцией создания
	pool := NewPoolResetableStruct(func() *ResetStruct {
		return &ResetStruct{Value: 1, Name: "test"}
	})

	assert.NotNil(t, pool, "NewPoolResetableStruct() вернул nil")
	assert.NotNil(t, pool.pool.New, "pool.pool.New не инициализирован")

	obj := pool.Get()
	assert.NotNil(t, obj, "Get() вернул nil")
	assert.Equal(t, 1, obj.Value)
	assert.Equal(t, "test", obj.Name)

	obj.Value = 2
	obj.Name = "modified"

	assert.Equal(t, 2, obj.Value)
	assert.Equal(t, "modified", obj.Name)

	objPitr := &obj

	pool.Put(obj)

	//Put должен вызвать reset, посмотрим через указатель на значения
	assert.Equal(t, 0, (*objPitr).Value)
	assert.Equal(t, "", (*objPitr).Name)

	//Get сохраненного объекта
	obj2 := pool.Get()
	assert.NotNil(t, obj2, "Get() вернул nil")
	assert.Equal(t, 0, obj2.Value)
	assert.Equal(t, "", obj2.Name)

	//Get нового объекта значения, должны быть значения из функции new
	obj3 := pool.Get()
	assert.NotNil(t, obj3, "Get() вернул nil")
	assert.Equal(t, 1, obj3.Value)
	assert.Equal(t, "test", obj3.Name)
}
