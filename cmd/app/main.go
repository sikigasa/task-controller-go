package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/sikigasa/task-controller/cmd/config"
	"github.com/sikigasa/task-controller/internal/infra"
	postgres "github.com/sikigasa/task-controller/internal/infra/driver"
	"github.com/sikigasa/task-controller/internal/usecase"
	task "github.com/sikigasa/task-controller/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func init() {
	config.LoadEnv(".env")
}

func main() {
	// 8080番portのListenerを作成
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := postgres.NewPostgresConnection(config.Config.Postgres.User, config.Config.Postgres.Password, config.Config.Postgres.Host, config.Config.Postgres.Port, config.Config.Postgres.DBName, config.Config.Postgres.SSLMode)
	if err != nil {
		panic(err)
	}
	db, err := conn.Connection()
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	// gRPCサーバーを作成
	s := grpc.NewServer()
	task.RegisterTaskServiceServer(s, usecase.NewTaskService(infra.NewTaskRepo(db), infra.NewTagRepo(db), infra.NewTaskTagRepo(db), postgres.NewPostgresTransaction(db)))
	task.RegisterTagServiceServer(s, usecase.NewTagService(infra.NewTagRepo(db)))

	reflection.Register(s)
	// 作成したgRPCサーバーを、8080番ポートで稼働させる
	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	// Ctrl+Cが入力されたらGraceful shutdownされるようにする
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}
