package pool

import (
	"sync"
)

/*
Более формально: вам необходимо реализовать структуру с generic-параметром, которая сможет хранить в себе объекты одного конкретного типа. \
Более того, такие объекты должны обладать методом Reset(), о котором мы рассказывали в прошлом инкременте.
Итоговый файл должен содержать:

    Структуру Pool с generic-параметром, который ограничен типами с методом Reset();
    Функцию-конструктор New, которая создаёт и возвращает указатель на структуру Pool;
    Метод Get() структуры Pool, который возвращает объект из пула;
    Метод Put() структуры Pool, который помещает объект в пул.
*/

// Объявлеем интерфейс для ограничения параметризированной структуры
type ResetableStruct interface {
	Reset()
}

// Пул сбрасываемых структур
type PoolResetableStruct[T ResetableStruct] struct {
	pool *sync.Pool
}

// Конструктро пула
func NewPoolResetableStruct[T ResetableStruct](NewResetableStruct func() T) *PoolResetableStruct[T] {
	pool := sync.Pool{
		New: func() any {
			return NewResetableStruct()
		},
	}
	return &PoolResetableStruct[T]{pool: &pool}
}

// GET метод пула
func (prs *PoolResetableStruct[T]) Get() T {
	rs := prs.pool.Get().(T)
	return rs
}

// PUT метод пула
func (prs *PoolResetableStruct[T]) Put(rs T) {
	rs.Reset()
	prs.pool.Put(rs)
}
