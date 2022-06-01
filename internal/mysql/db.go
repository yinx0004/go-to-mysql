package mysql

import (
	"context"
	"database/sql"
	"github.com/rs/zerolog/log"
	"go-to-mysql/internal"
	"go-to-mysql/internal/rand"
	"time"
)

const (
	stmtTimeout = 10 * time.Second
)

type Conn struct {
	DB *sql.DB
}

func (c Conn) CreateDB(dbName string) error {
	stmt := "create database if not exists " + dbName
	ctx, cancel := context.WithTimeout(context.Background(), stmtTimeout)
	defer cancel()
	if _, err := c.DB.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func (c Conn) CreateTab(dbName string) error {
	stmt := "create table if not exists " + dbName + ".test_tab" + " (id int not null auto_increment primary key,col1 int not null, col2 char(20) not null)"
	ctx, cancel := context.WithTimeout(context.Background(), stmtTimeout)
	defer cancel()
	if _, err := c.DB.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func (c Conn) Insert(dbName string) error {
	funcName := internal.GetFuncName()
	log.Debug().Str("func", funcName).Msg("")
	col1 := rand.Int(100000000)
	col2 := rand.String(20)
	stmt := "insert into " + dbName + ".test_tab (id, col1, col2) values (0, ?, ?)"
	if _, err := c.DB.Exec(stmt, col1, col2); err != nil {
		return err
	}
	return nil
}
