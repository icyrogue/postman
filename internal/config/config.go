package config

import (
	"flag"
	"os"

	"github.com/icyrogue/postman/internal/analyzer"
	"github.com/icyrogue/postman/internal/asyncstorage"
	"github.com/icyrogue/postman/internal/checker"
	"github.com/icyrogue/postman/internal/client"
	"github.com/icyrogue/postman/internal/dbstorage"
)

type config struct {
	ClientOpts       *client.Options
	DBstorageOpts    *dbstorage.Options
	AnalyzerOpts     *analyzer.Options
	AsyncStorageOpts *asyncstorage.Options
	CheckerOpts      *checker.Options
}

func New() *config {
	return &config{
		ClientOpts:       &client.Options{},
		DBstorageOpts:    &dbstorage.Options{},
		AnalyzerOpts:     &analyzer.Options{},
		AsyncStorageOpts: &asyncstorage.Options{},
		CheckerOpts:      &checker.Options{},
	}
}

func (cfg *config) Get() error {
	flag.StringVar(&cfg.DBstorageOpts.DSN, "d", "", "путь к БД")
	flag.IntVar(&cfg.AsyncStorageOpts.MaxWaitTime, "t", 60, "Макс время до обращения к БД")
	flag.IntVar(&cfg.AsyncStorageOpts.MaxBufferLength, "b", 10, "Макс размер очереди записей до обращения к БД")
	flag.IntVar(&cfg.CheckerOpts.WaitTime, "c", 60, "Макс время до повторной проверки новых уведолений")
	flag.StringVar(&cfg.AnalyzerOpts.TemplatesDir, "template", "templates/", " Путь до шаблонов для рассылки")

	if err := flag.Lookup("d").Value.Set(os.Getenv("DATABASE_DSN")); err != nil {
		return err
	}

	cfg.ClientOpts.Server = os.Getenv("POSTMAN_SERVER")
	cfg.ClientOpts.Port = os.Getenv("POSTMAN_PORT")
	cfg.ClientOpts.Password = os.Getenv("POSTMAN_PASSWORD")
	cfg.ClientOpts.Login = os.Getenv("POSTMAN_LOGIN")

	cfg.CheckerOpts.Server = os.Getenv("POSTMAN_IMAP_SERVER")
	cfg.CheckerOpts.Port = os.Getenv("POSTMAN_IMAP_PORT")
	cfg.CheckerOpts.Login = cfg.ClientOpts.Login
	cfg.CheckerOpts.Password = cfg.ClientOpts.Password

	return nil

}
