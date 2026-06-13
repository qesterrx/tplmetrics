// Пакет retry - пакет содержит функцию для повтора операции по заданным условиям
package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
)

// RetryableFunc - тип описывающий операцию в виде функции возвращающей ошибку
// Успешным выполнением операции считается nil в результате выполнения функции
type RetryableFunc func() error

// CheckRetryableFunc - тип описывающий функцию для проверки ошибки на необходимость выполнения повтора
type CheckRetryableFunc func(error) bool

// RetryFunc - функция повторения операции
//
// На вход ожидает следующие параметры
// ctx context.Context - Контекст выполнения для отмены
// function RetryableFunc - Функция описывающая операцию
// checker CheckRetryableFunc - Функция проверяющая ошибку на необходимость переотправки
// countAttempt int - Количество попыток
// startDelay time.Duration - Задержка перед первым повтором выполнения
// iterationDelay time.Duration - Задержка между следующими повторами выполнения
//
// На выходе одтает nil или финальную ошибку выполнения
func RetryFunc(ctx context.Context, function RetryableFunc, checker CheckRetryableFunc, countAttempt int, startDelay time.Duration, iterationDelay time.Duration) error {

	//Мало ли
	if countAttempt == 1 {
		return function()
	}

	var err error

	delay := startDelay

	for attempt := 1; attempt <= countAttempt; attempt++ {

		err = function()
		retry := false

		if err != nil {
			logger.Log.Debug().Msg(fmt.Sprintf("Ошибка выполнения retryableFunc, попытка %d: %s \n", attempt, err.Error()))

			//Проверить
			if checker(err) {
				retry = true
			}

		} else {
			//Выполнено успешно
			return nil
		}

		//Если надо ждать то чекнем что это не последняя попытка
		if retry && attempt != countAttempt {
			select {
			case <-ctx.Done():
				return err
			case <-time.After(delay):
				delay = iterationDelay
			}
		} else {
			return err
		}

	}

	//Попытки закончились
	return err
}
