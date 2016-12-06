package db

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init(dbpath string) error {
	err := os.MkdirAll(path.Dir(dbpath), 0755)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(dbpath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	f.Close()
	d, err := sql.Open("sqlite3", dbpath+"?charset=utf8&sql_mode=ANSI_QUOTES")
	if err != nil {
		return err
	}
	db = d
	return checkTables()
}

func Get() *sql.DB {
	if db == nil {
		panic(fmt.Errorf("No avaliable database"))
	}
	return db
}

func RangeRows(rows *sql.Rows, ds func() error) error {
	defer func() {
		if err := recover(); err != nil {
			rows.Close()
			panic(err)
		}
	}()
	for rows.Next() {
		if err := ds(); err != nil {
			rows.Close()
			return err
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	return nil
}
