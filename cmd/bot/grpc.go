package main

import (
	"context"
	"fmt"
	"net"

	"gitlab.ozon.dev/cranky4/tg-bot/api"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicemessages "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	api.UnimplementedReporterServer
	messagesService *servicemessages.Model
}

func (s *server) SendReport(ctx context.Context, request *api.SendReportRequest) (*emptypb.Empty, error) {
	err := s.messagesService.SendReport(&expense_reporter.ExpenseReport{
		IsEmpty: request.IsEmpty,
		Rows:    request.Rows,
		UserID:  request.UserId,
	})
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func initGRPСServer(conf config.GRPCConf, messagesService *servicemessages.Model) error {
	port := fmt.Sprintf(":%d", conf.Port)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	api.RegisterReporterServer(s, &server{messagesService: messagesService})

	logger.Info("GRPC server listening " + port)

	if err := s.Serve(listener); err != nil {
		return err
	}

	return nil
}
