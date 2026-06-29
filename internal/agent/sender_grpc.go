package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gookit/goutil/netutil"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func SenderGRPC(ctx context.Context, toSend <-chan []byte, url string) {
	logger.Log.Debug().Msg("Запуск GRPC Sender")
	//Клиента создаем один раз
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Error().Msg("Оошибка при установлении соединения с сервером")
		return
	}
	defer conn.Close()
	client := proto.NewMetricsClient(conn)

	//Тут определим middleware агента
	ipAgnet := netutil.InternalIPv4()
	md := metadata.Pairs("x-real-ip", ipAgnet)

	ctxClient := metadata.NewOutgoingContext(ctx, md)

	send := func(msg []byte) error {
		//Десериализация
		mtrks := []model.MetricaJSONAdapter{}
		err := json.Unmarshal(msg, &mtrks)
		if err != nil {
			return fmt.Errorf("Ошибка десериализации метрик: %w", err)
		}

		gms := []*proto.Metric{}
		for _, mtrk := range mtrks {
			gm := proto.Metric{}
			gm.SetId(mtrk.Name)
			if model.KindValue(mtrk.Kind) == model.Counter {
				gm.SetType(proto.Metric_COUNTER)
				gm.SetDelta(*mtrk.Delta)
			} else {
				gm.SetType(proto.Metric_GAUGE)
				gm.SetValue(*mtrk.Value)
			}
			gms = append(gms, &gm)
		}

		//Запрос
		req := proto.UpdateMetricsRequest_builder{Metrics: gms}

		//Отправка
		_, err = client.UpdateMetrics(ctxClient, req.Build())
		if err != nil {
			return fmt.Errorf("Ошибка отправки метрик через GRPC : %w", err)
		}

		logger.Log.Debug().Msg("Метрики отправлены через через GRPC")
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			//Reporter может еще что-то дописывать в канал, ждем закрытия канала
			logger.Log.Debug().Msg("Остановка GRPC Sender по контексту, ожидается завершение процесса")

			for msg := range toSend {
				err := send(msg)
				if err != nil {
					logger.Log.Error().Msg(err.Error())
				}
			}
			//Если получили сигнал завершения останавливаемся
			return
		case msg := <-toSend:
			err := send(msg)
			if err != nil {
				logger.Log.Error().Msg(err.Error())
			}
		}
	}
}
