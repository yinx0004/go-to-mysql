package mysql

import (
	"context"
	"database/sql"
	"github.com/rs/zerolog/log"
	"go-to-mysql/internal"
	"time"
)

const (
	DDLTimeout = 10 * time.Second
)

type Conn struct {
	DB *sql.DB
}

func (c Conn) CreateDB(dbName string) error {
	stmt := "create database if not exists " + dbName
	log.Debug().Str("create database stmt:", stmt).Msg("")
	ctx, cancel := context.WithTimeout(context.Background(), DDLTimeout)
	defer cancel()
	if _, err := c.DB.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func (c Conn) CreateTab(dbName string) error {
	stmt := "create table if not exists " + dbName + ".test_tab" + " (id int not null auto_increment primary key,col1 int not null, col2 char(20) not null)"
	ctx, cancel := context.WithTimeout(context.Background(), DDLTimeout)
	defer cancel()
	if _, err := c.DB.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func (c Conn) Insert(dbName string, col1 int, col2 string) error {
	defer internal.TimeTrack(time.Now())
	funcName := internal.GetFuncName()
	log.Debug().Str("func", funcName).Msg("")
	master, err := c.GetSysVar("wsrep_node_name")
	if err != nil {
		return err
	}
	log.Debug().Str("master", *master).Msg("Current master")
	stmt := "insert into " + dbName + ".test_tab (id, col1, col2) values (0, ?, ?)"
	if _, err := c.DB.Exec(stmt, col1, col2); err != nil {
		return err
	}
	return nil
}

func (c Conn) GetSysVar(sysvar string) (*string, error) {
	stmt := "show global variables like " + "'" + sysvar + "'"
	res := c.DB.QueryRow(stmt)
	var key *string
	var value *string
	if err := res.Scan(&key, &value); err != nil {
		return nil, err
	}
	return value, nil
}
