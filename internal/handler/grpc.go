package handler

import (
	"context"
	"fmt"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	pb "github.com/qesterrx/tplmetrics/internal/proto"
	"github.com/qesterrx/tplmetrics/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCContainer struct {
	pb.UnimplementedMetricsServer
	tcl *service.TCLService
}

func NewGRPCContainer(tcl *service.TCLService) *GRPCContainer {
	return &GRPCContainer{tcl: tcl}
}

func (grpch *GRPCContainer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {

	var resp pb.UpdateMetricsResponse
	mtrks := []model.Metrica{}

	for _, reqm := range req.GetMetrics() {
		reqmType := reqm.GetType()
		switch reqmType {
		case pb.Metric_COUNTER:
			mtrk := model.NewMetricaCounter(reqm.GetId(), reqm.GetDelta())
			mtrks = append(mtrks, mtrk)
		case pb.Metric_GAUGE:
			mtrk := model.NewMetricaGauge(reqm.GetId(), reqm.GetValue())
			mtrks = append(mtrks, mtrk)
		default:
			return &resp, status.Error(codes.Internal, "Неподдерживаемый тип метрики")
		}
	}

	err := grpch.tcl.UpdateMetricaBatch(ctx, mtrks)
	if err != nil {
		logger.Log.Info().Msg(fmt.Sprintf("GRPC UpdateMetrics ошибка при обновлении метрик: %s", err.Error()))
		return &resp, status.Error(codes.Internal, err.Error())
	}

	return &resp, nil
}
