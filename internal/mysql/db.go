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
	log.Debug().Str("statement:", stmt).Msg("")
	ctx, cancel := context.WithTimeout(context.Background(), DDLTimeout)
	defer cancel()
	if _, err := c.DB.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func (c Conn) CreateTab(dbName string) error {
	stmt := "create table if not exists " + dbName + ".test_tab" + " (id int not null auto_increment primary key,col1 int not null, col2 char(20) not null)"
	log.Debug().Str("statement", stmt).Msg("")
	ctx, cancel := context.WithTimeout(context.Background(), DDLTimeout)
	defer cancel()
	if _, err := c.DB.ExecContext(ctx, stmt); err != nil {
		return err
	}
	return nil
}

func (c Conn) Insert(dbName string, col1 int, col2 string) (*string, error) {
	defer internal.TimeTrack(time.Now())
	funcName := internal.GetFuncName()
	log.Debug().Str("func", funcName).Msg("")
	master, err := c.GetSysVar("wsrep_node_name")
	if err != nil {
		return nil, err
	}
	stmt := "insert into " + dbName + ".test_tab (id, col1, col2) values (0, ?, ?)"
	log.Debug().Str("statement", stmt).Msg("")
	if _, err := c.DB.Exec(stmt, col1, col2); err != nil {
		return nil, err
	}
	return master, nil
}

func (c Conn) Txn(dbName string, col1 int, col2 string) (*string, error) {
	defer internal.TimeTrack(time.Now())

	master, err := c.GetSysVar("wsrep_node_name")
	if err != nil {
		return nil, err
	}

	tx, err := c.DB.Begin()
	if err != nil {
		return nil, err
	}
	stmt1 := "insert into " + dbName + ".test_tab (id, col1, col2) values (0, ?, ?)"
	log.Debug().Str("statement", stmt1).Msg("")
	x, err := tx.Exec(stmt1, col1, col2)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	genID, err := x.LastInsertId()
	if err != nil {
		return nil, err
	}
	stmt2 := "select col1, col2 from " + dbName + ".test_tab where id = ?"
	log.Debug().Str("statement", stmt2).Msg("")
	res := tx.QueryRow(stmt2, genID)
	var rescol1 *int
	var rescol2 *string
	if err := res.Scan(&rescol1, &rescol2); err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	log.Debug().Int64("id", genID).Int("col1", *rescol1).Str("col2", *rescol2).Msg("")
	return master, nil
}

func (c Conn) GetSysVar(sysvar string) (*string, error) {
	stmt := "show global variables like " + "'" + sysvar + "'"
	log.Debug().Str("statement", stmt).Msg("")
	res := c.DB.QueryRow(stmt)
	var key *string
	var value *string
	if err := res.Scan(&key, &value); err != nil {
		return nil, err
	}
	return value, nil
}
