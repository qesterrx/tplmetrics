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

type ResetableStruct interface {
	Reset()
}

type PoolResetableStruct[T ResetableStruct] struct {
	pool *sync.Pool
}

func NewPoolResetableStruct[T ResetableStruct](NewResetableStruct func() *T) *PoolResetableStruct[T] {
	pool := sync.Pool{
		New: func() interface{} {
			return NewResetableStruct()
		},
	}
	return &PoolResetableStruct[T]{pool: &pool}
}

func (prs *PoolResetableStruct[T]) Get() T {
	rs := prs.pool.Get().(T)
	return rs
}

func (prs *PoolResetableStruct[T]) Put(rs T) {
	rs.Reset()
	prs.pool.Put(rs)
}
