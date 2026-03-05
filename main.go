package main

import (
	"context"
	"fmt"
	"os"
	httpServer "test-assignment/internal/http"
	"test-assignment/internal/storage/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := postgres.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer storage.Close(ctx)

	if err := storage.Ping(ctx); err != nil {
		panic(err)
	}

	handlers := httpServer.NewHTTPHandlers(storage)

	server := httpServer.NewHTTPServer(handlers)

	fmt.Println("SERVER_ADDRESS\t", os.Getenv("SERVER_ADDRESS"))
	server.StartServer(os.Getenv("SERVER_ADDRESS"))
}
