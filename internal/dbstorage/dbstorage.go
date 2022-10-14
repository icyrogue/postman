package dbstorage

import (
	"context"
	"errors"
	"log"
	"net/mail"

	"github.com/jackc/pgx/v5"
)

type storage struct {
	conn    *pgx.Conn
	Options *Options
}

type Options struct {
	DSN string
}

func New() *storage {
	return &storage{}
}

//Init: подулючается к БД по Options.DSN, создает новую таблицу
func (st *storage) Init() error {
	conn, err := pgx.Connect(context.Background(), st.Options.DSN)
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS "users" (id TEXT, info JSONB, readtime TEXT)`)
	if err != nil {
		return err
	}

	st.conn = conn
	return nil
}

//Close: вежливо благодарим базу данных за проделанную работу
func (st *storage) Close() error {
	return st.conn.Close(context.Background())
}

//Ping: проверка соединения с базой данных
func (st *storage) Ping(ctx context.Context) error {
	return st.conn.Ping(ctx)
}

//NewList: проверяем, есть ли уже в БД лист рассылки с таким  id,
//если есть, то возвращаем ошибку
func (st *storage) NewList(ctx context.Context, id string) error {
	var count int8
	err := st.conn.QueryRow(ctx, `SELECT COUNT(*) FROM "users" where id = $1`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("id already exists")
	}
	return nil
}

//Add: добавить все новые записи
func (st *storage) Add(ctx context.Context, data [][]interface{}) error {
	names := []string{"id", "info"}
	_, err := st.conn.CopyFrom(ctx, pgx.Identifier{"users"}, names, pgx.CopyFromRows(data))
	if err != nil {
		return err
	}
	return nil
}

//Get: получить полный список пользователей в рассылке с id
func (st *storage) Get(ctx context.Context, id string) (users []byte, error error) {
	err := st.conn.QueryRow(ctx, `SELECT jsonb_agg(c.info) FROM (SELECT info FROM "users" WHERE id = $1) c`, id).Scan(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

//AddRead: добавить время и пользователя, который прочитал email
func (st *storage) AddRead(ctx context.Context, addr, tmStamp string) error {
	email, err := mail.ParseAddress(addr)
	if err != nil {
		return err
	}
	log.Println(email.Address)
	_, err = st.conn.Exec(ctx, `UPDATE "users" SET readtime = $1 WHERE info->>'email' LIKE $2 `, tmStamp, email.Address)
	if err != nil {
		return err
	}
	return nil
}

//GetRead: получить статистику о прочтении письма
func (st *storage) GetRead(ctx context.Context, id string) (users []byte, err error) {
	err = st.conn.QueryRow(ctx, `SELECT json_agg(c.json_build_object) FROM (SELECT json_build_object('email', info ->> 'email', 'time', readtime) FROM users WHERE id = $1) c`, id).Scan(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}
