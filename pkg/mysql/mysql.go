package mysql

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func Connect(dsn string) (*DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &DB{db}, nil
}

func (db *DB) NamedSelect(dest any, query string, arg any) error {
	q, args, err := sqlx.Named(query, arg)
	if err != nil {
		return err
	}
	q = db.Rebind(q)
	return db.Select(dest, q, args...)
}

func (db *DB) NamedGet(dest any, query string, arg any) error {
	q, args, err := sqlx.Named(query, arg)
	if err != nil {
		return err
	}
	q = db.Rebind(q)
	return db.Get(dest, q, args...)
}
