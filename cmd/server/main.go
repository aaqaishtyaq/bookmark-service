package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	v1pb "github.com/aaqaishtyaq/bookmark-service/pkg/api/v1"
	"github.com/aaqaishtyaq/bookmark-service/pkg/logger"
	v1 "github.com/aaqaishtyaq/bookmark-service/pkg/service/v1"
	_ "github.com/mattn/go-sqlite3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := runServer(); err != nil {
		log.Printf("%+v\n", err)
		os.Exit(1)
	}
}

const (
	dbName string = "bookmark.db"

	schema string = `
CREATE TABLE IF NOT EXISTS bookmarks (
id INTEGER NOT NULL PRIMARY KEY,
url TEXT NOT NULL,
Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);`
)

// Config is configuration for Server
type Config struct {
	// gRPC server start parameters section
	// GRPCPort is TCP port to listen by gRPC server
	GRPCPort string
	// Log parameters section
	// LogLevel is global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	LogLevel int
	// LogTimeFormat is print time format for logger e.g. 2006-01-02T15:04:05Z07:00
	LogTimeFormat string
}

func runServer() error {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get configuration
	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.IntVar(&cfg.LogLevel, "log-level", 0, "Global log level")
	flag.StringVar(&cfg.LogTimeFormat, "log-time-format", "",
		"Print time format for logger e.g. 2006-01-02T15:04:05Z07:00")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	// initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.LogTimeFormat); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return err
	}

	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// slice of gRPC options
	// Here we can configure things like TLS
	opts := []grpc.ServerOption{}
	// var s *grpc.Server
	s := grpc.NewServer(opts...)

	v1API := v1.NewBookmarkServiceServer(db)
	v1pb.RegisterBookmarkServiceServer(s, v1API)
	reflection.Register(s)

	listen, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return err
	}

	go func() {
		if err := s.Serve(listen); err != nil {
			logger.Log.Sugar().Errorf("Failed to serve: %v", err)
		}
	}()

	logger.Log.Sugar().Info("Server succesfully started on port :50051")

	// Create a channel to receive OS signals
	c := make(chan os.Signal, 1)

	// Relay os.Interrupt to our channel (os.Interrupt = CTRL+C)
	// Ignore other incoming signals
	signal.Notify(c, os.Interrupt)
	<-c

	// After receiving CTRL+C Properly stop the server
	log.Printf("\nStopping the server...")
	s.Stop()
	listen.Close()
	logger.Log.Sugar().Info("Done.")
	return nil
}
