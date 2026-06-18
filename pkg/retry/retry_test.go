package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {

	var cnt int
	var err error

	//Чекеры ошибок
	chTrue := func(error) bool { return true }
	chFalse := func(error) bool { return false }

	//Функция fn возвращающая ошибку
	fn := func() error {
		cnt++
		return errors.New("test")
	}

	//Повторы
	cnt = 0
	start := time.Now()
	err = RetryFunc(t.Context(), fn, chTrue, 3, 10*time.Millisecond, 20*time.Millisecond)

	assert.WithinDuration(t, start.Add((10+20)*time.Millisecond), time.Now(), (10+20+10)*time.Millisecond) //даем 10 мс на выполнение fn
	assert.Error(t, err)
	assert.Equal(t, 3, cnt, "Количество вызовов retry не соответсвует ожидаемомоу")

	//Нет повторов
	cnt = 0
	start = time.Now()
	err = RetryFunc(t.Context(), fn, chFalse, 2, 10*time.Millisecond, 20*time.Millisecond)

	assert.WithinDuration(t, start, time.Now(), 10*time.Millisecond) //даем 10 мс на выполнение fn
	assert.Error(t, err)
	assert.Equal(t, 1, cnt, "Количество вызовов retry не соответсвует ожидаемомоу")

	//Функция fn возвращающая успех
	fn = func() error {
		cnt++
		return nil
	}
	//Повторы
	cnt = 0
	start = time.Now()
	err = RetryFunc(t.Context(), fn, chTrue, 3, 10*time.Millisecond, 20*time.Millisecond)

	assert.WithinDuration(t, start, time.Now(), 10*time.Millisecond) //даем 10 мс на выполнение fn
	assert.NoError(t, err)
	assert.Equal(t, 1, cnt, "Количество вызовов retry не соответсвует ожидаемомоу")

	//Нет повторов
	cnt = 0
	start = time.Now()
	err = RetryFunc(t.Context(), fn, chFalse, 2, 10*time.Millisecond, 20*time.Millisecond)

	assert.WithinDuration(t, start, time.Now(), 10*time.Millisecond) //даем 10 мс на выполнение fn
	assert.NoError(t, err)
	assert.Equal(t, 1, cnt, "Количество вызовов retry не соответсвует ожидаемомоу")

}

func ExampleRetryFunc() {

	//Чекеры ошибок
	chTrue := func(error) bool { return true }

	//Функция fn возвращающая ошибку
	fn := func() error {
		fmt.Println("run fn", time.Now())
		return errors.New("test error")
	}

	//В данном случае функция fn будет выполняться 3 раза, первый раз сразу, второ раз через 10ms и еще один раз через 20ms
	RetryFunc(context.Background(), fn, chTrue, 3, 10*time.Millisecond, 20*time.Millisecond)

}
