package main

import (
	"context"
	"log"

	"github.com/icyrogue/postman/internal/analyzer"
	"github.com/icyrogue/postman/internal/api"
	"github.com/icyrogue/postman/internal/asyncstorage"
	"github.com/icyrogue/postman/internal/checker"
	"github.com/icyrogue/postman/internal/client"
	"github.com/icyrogue/postman/internal/config"
	"github.com/icyrogue/postman/internal/dbstorage"
	"github.com/icyrogue/postman/internal/requestprocessor"
)

func main() {
	config := config.New()
	if err := config.Get(); err != nil {
		log.Fatal(err.Error())
	}

	storage := dbstorage.New()
	storage.Options = config.DBstorageOpts
	defer storage.Close()
	if err := storage.Init(); err != nil {
		log.Fatal(err.Error())
	}

	mailClient := client.New()
	mailClient.Options = config.ClientOpts
	defer client.New().Close()
	if err := mailClient.Init(); err != nil {
		log.Fatal(err.Error())
	}

	asyncStorage := asyncstorage.New(storage)
	asyncStorage.Options = config.AsyncStorageOpts
	asyncStorage.Start(context.Background())

	requestProcessor := requestprocessor.New(storage, asyncStorage)

	mailAnalyzer := analyzer.New(storage, mailClient)
	mailAnalyzer.Options = config.AnalyzerOpts
	defer mailAnalyzer.Stop()
	mailAnalyzer.Init()

	mailChecker := checker.New(storage)
	mailChecker.Options = config.CheckerOpts
	defer mailChecker.Close()
	if err := mailChecker.Init(); err != nil {
		log.Println(err.Error())
	}
	mailChecker.Start(context.Background())

	api := api.New(storage, requestProcessor, mailAnalyzer)
	api.Init()
	api.Run()
}
