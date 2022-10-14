package requestprocessor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"time"
)

type requestprocessor struct {
	storage      Storage
	asyncStorage AsyncStorage
}

type Storage interface {
	NewList(ctx context.Context, id string) error
}

type AsyncStorage interface {
	GetInput() chan []interface{}
}

func New(storage Storage, asyncStorage AsyncStorage) *requestprocessor {
	return &requestprocessor{storage: storage, asyncStorage: asyncStorage}
}

//init: берет сид для случайных чисел
func init() {
	rand.Seed(time.Now().UnixMicro())
}

//genID:  генерирует id из 8 случайных чисел
func genID() string {
	chars := []byte("qwertyuiopasdfghklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")
	output := []byte{}
	for i := 0; i != 8; i++ {
		output = append(output, chars[rand.Intn(len(chars))])
	}
	return string(output)
}

//NewList: создает новый список со случайным id
func (r *requestprocessor) NewList(ctx context.Context) (id string, err error) {
	for {
		id = genID()
		if err = r.storage.NewList(ctx, id); err == nil { //если БД не вернула ошибку, значить такого списка еще нет
			return id, nil
		}
		if errors.Is(err, errors.New("id already exists")) {
			continue
		}
		return "", err
	}
}

//AddUser: добавляет пользователя в список рассылки по id
func (r *requestprocessor) AddUser(id string, data []byte) error {
	if ok := json.Valid(data); !ok {
		return errors.New("invalid JSON")
	}
	//БД ожидает новую запись в формате списка, где первый элемент - это ID, второй -  JSON модель пользователя
	row := make([]interface{}, 2)
	row[0] = id
	row[1] = data
	r.asyncStorage.GetInput() <- row
	return nil
}

//AddBatch: читает список пользователей в JSON, скидывет каждую запись в очередь
func (r *requestprocessor) AddBatch(id string, data []byte) error {
	if ok := json.Valid(data); !ok {
		return errors.New("invalid JSON array")
	}
	first := true //нужно, чтобы избавиться от [ в первом элементе списка
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '}'); i >= 0 {
			i++
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(split)
	go func() {
		for scanner.Scan() {
			body := scanner.Bytes()
			if first {
				i := bytes.IndexByte(body, '[')
				body = body[i+1:]
			}
			if len(body) == 1 {
				continue
			}
			log.Println(string(body))
			user := make([]interface{}, 2)
			user[0] = id
			user[1] = body

			r.asyncStorage.GetInput() <- user
		}
	}()
	return nil
}
