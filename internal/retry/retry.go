package retry

import (
	"context"
	"fmt"
	"time"
)

type RetryableFunc func() error
type CheckRetryableFunc func(error) bool

func RetryFunc(ctx context.Context, function RetryableFunc, checker CheckRetryableFunc, countAttempt int, startDelay time.Duration, iterationDelay time.Duration) error {

	var err error

	delay := startDelay

	for attempt := 1; attempt <= countAttempt; attempt++ {

		err = function()
		retry := false

		if err != nil {
			fmt.Printf("Ошибка выполнения retryableFunc, попытка %d: %s \n", attempt, err.Error()) //Временный вариант

			//Проверить
			if checker(err) {
				retry = true
			}

		} else {
			//Выполнено успешно
			return nil
		}

		//Ожидаем
		if retry {
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
