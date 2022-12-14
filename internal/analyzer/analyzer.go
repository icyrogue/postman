package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"os"
	"path"

	"github.com/gocelery/gocelery"
	"github.com/gomodule/redigo/redis"
)

type analyzer struct {
	celery  *gocelery.CeleryClient
	storage Storage
	client  Client
	Options *Options
}

type Options struct {
	TemplatesDir string
}

type worker struct {
	id       string
	data     map[string]interface{}
	template *template.Template
	client   Client
}

type Storage interface {
	Get(ctx context.Context, id string) (data []byte, err error)
}

type Client interface {
	Send(id string, addr string, body []byte) error
}

func New(storage Storage, client Client) *analyzer {
	return &analyzer{storage: storage, client: client}
}

//Init: создает новый пул для брокера Redis, добавляет обработчика Celery
func (a *analyzer) Init() {
	rdPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL("redis://")
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	cli, _ := gocelery.NewCeleryClient(
		gocelery.NewRedisBroker(rdPool),
		&gocelery.RedisCeleryBackend{Pool: rdPool},
		5,
	)
	cli.Register("postman.mail", a.mail)
	cli.StartWorker()
	a.celery = cli
}

//Stop: отпускаем рабочих домой
func (a *analyzer) Stop() {
	a.celery.StartWorker()
}

//StoreTemplate: записывает шаблон для рассылки, путь для сохранения указан в Options.TemplatesDir,
//для имени файла в формате html использует id рассылки
func (a *analyzer) StoreTemplate(id string, data []byte) error {
	err := os.WriteFile(path.Join(a.Options.TemplatesDir, id+".html"), data, 0644)
	if err != nil {
		return err
	}
	return nil
}

//mail: простой обработчик Celery: отправляет письма всем пользователям из списка с полученным idб
//возвращает кол-во пользователей в рассылке или 0, если при рассылке произошла ошибка
func (a *analyzer) mail(id string) int {
	body, err := template.ParseFiles(path.Join(a.Options.TemplatesDir, id+".html"))
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	data, err := a.storage.Get(context.Background(), id)
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	log.Println(string(data))
	var users []interface{}
	err = json.Unmarshal(data, &users)
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	go func() {
		for _, usr := range users {
			if usr == nil {
				continue
			}
			w := worker{data: usr.(map[string]interface{}), template: body, client: a.client}
			if err := w.do(); err != nil {
				log.Println(err.Error())
			}
		}
	}()
	return len(users)
}

//do: загоняет информацию о пользователе в шаблон, отправлет письмо через клиент
func (w *worker) do() error {
	var body bytes.Buffer
	err := w.template.Execute(&body, w.data)
	if err != nil {
		return err
	}
	email, ok := w.data["email"]
	if !ok {
		return errors.New("couldn't find user email")
	}
	if email == nil {
		return errors.New("couldn't find user email")
	}
	err = w.client.Send(w.id, email.(string), body.Bytes())
	if err != nil {
		return err
	}
	return nil
}
