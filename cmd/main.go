package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go-to-mysql/internal/mysql"
)

const (
	pingTimeout = 5 * time.Second
	connTimeout = 10 * time.Second
	sleepTime   = 5 * time.Second
)

type config struct {
	concurrency int
	debug       bool
	dbName      string
	dsn         struct {
		host      string
		port      string
		user      string
		passwd    string
		parseTime string
		timeout   int
	}
}

type application struct {
	db  mysql.Conn
	log zerolog.Logger
}

func main() {
	var cfg config

	flag.StringVar(&cfg.dsn.host, "h", "localhost", "MySQL host")
	flag.StringVar(&cfg.dsn.port, "P", "3306", "MySQL server port")
	flag.StringVar(&cfg.dsn.user, "u", "root", "MySQL user")
	flag.StringVar(&cfg.dsn.passwd, "p", "", "MySQL password")
	flag.StringVar(&cfg.dsn.parseTime, "T", "true", "MySQL parseTime(true|false)")
	flag.IntVar(&cfg.concurrency, "c", 50, "Number of Goroutione")
	flag.BoolVar(&cfg.debug, "debug", false, "show debug level log")
	flag.StringVar(&cfg.dbName, "d", "", "MySQL database name")
	flag.Parse()

	log.Info().Msg("Starting program...")
	dsn := getDSN(cfg)
	conn, err := initDB(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("DB connection pool initialization failed.")
	}
	defer conn.Close()
	log.Info().Msg("DB connection pool initialized.")

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	app := application{
		db:  mysql.Conn{DB: conn},
		log: logger,
	}

	app.log.With().Timestamp()

	if err = app.db.CreateDB(cfg.dbName); err != nil {
		app.log.Fatal().Err(err).Msg("Database init failed.")
	}
	if err = app.db.CreateTab(cfg.dbName); err != nil {
		app.log.Fatal().Err(err).Msg("Create test table failed.")
	}
	app.log.Info().Msg("Database initialization complete.")

	app.log.Info().Msg("Start to insert data...")
	for {
		var wg sync.WaitGroup
		for i := 0; i < cfg.concurrency; i++ {
			wg.Add(1)
			go func(i int) {
				err = app.db.Insert(cfg.dbName)
				if err != nil {
					app.log.Error().Err(err).Msg("Insert failed.")
				}
				app.log.Debug().Int("goroutine", i).Msg("Insert Succeed.")
				wg.Done()
			}(i)
		}
		wg.Wait()
		app.log.Info().Msg("in progress...")
		//time.Sleep(sleepTime)
	}
}

func initDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	conn.SetConnMaxLifetime(4 * time.Minute)
	conn.SetMaxOpenConns(50)
	conn.SetMaxIdleConns(50)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err = conn.PingContext(ctx); err != nil {
		return nil, err
	}
	return conn, nil
}

func getDSN(cfg config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=%s&timeout=%s", cfg.dsn.user, cfg.dsn.passwd, cfg.dsn.host, cfg.dsn.port, cfg.dsn.parseTime, connTimeout)
}
