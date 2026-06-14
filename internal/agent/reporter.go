package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

// Reporter Процедура (горутина) - через reportInterval вычитывает очередь toGroup, Gauge метрики группирует, Counter метрики оставляет последнюю из очереди. Далее ставит сгруппированные/отфильтрованные значения в очередь на отправку toSend в виде []byte
func Reporter(ctx context.Context, toGroup <-chan model.Metrica, toSend chan<- []byte, reportInterval int) {
	logger.Log.Debug().Msg("Запуск Reporter")

	ticker := time.NewTicker(time.Second * time.Duration(reportInterval))
	defer ticker.Stop()

	//в данной структуре будем группировать данные из очереди, кроме того с помощью нее обеспечим транзакционность
	groupMap := map[string]model.Metrica{}

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			logger.Log.Debug().Msg("Остановка Reporter по контексту")
			return
		case <-ticker.C:

			logger.Log.Debug().Msg("Reporter запуск группировки данных из очереди toGroup")

		loop:
			for {
				select {
				case <-ctx.Done():
					//Если получили сигнал завершения останавливаемся
					logger.Log.Debug().Msg("Остановка Reporter по контексту")
					return
				case metrica := <-toGroup:
					//Группировка
					switch metrica.Kind() {
					case model.Gauge:
						groupMap[metrica.Name()] = metrica
					case model.Counter:
						oldMetrica, ok := groupMap[metrica.Name()]
						if ok {
							err := oldMetrica.UpdateValueAtomic(metrica)
							if err != nil {
								logger.Log.Error().Msg("Ошибка обновления метрики " + err.Error())
								continue
							}
						} else {
							groupMap[metrica.Name()] = metrica
						}
					default:
						logger.Log.Error().Msg("Неизвестный тип метрики " + metrica.Name() + string(metrica.Kind()))
					}
				default:
					//Канал вычитан до конца - выходим из цикла
					break loop
				}
			}

			//Сериализация и ставим в очередь на отправку - Ну вот и пригодилось
			mtrks := []model.Metrica{}

			for _, metrica := range groupMap {
				mtrks = append(mtrks, metrica)
			}

			body, err := json.Marshal(mtrks)
			if err != nil {
				logger.Log.Error().Msg("Ошибка сериализации метрик " + err.Error())
				continue
			}

			logger.Log.Debug().Msg(fmt.Sprintf("Попытка отправить массив метрик, количество %d", len(mtrks)))

			//Отправка данных
			toSend <- body

			//... и каков ответ на главный вопрос жизни, вселенной и всего такого
			groupMap = map[string]model.Metrica{}

		}
	}

}
