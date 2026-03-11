package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
)

type RetryableFunc func() error
type CheckRetryableFunc func(error) bool

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
