package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"go-to-mysql/internal/random"
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
	version     = "1.0.1"
)

type config struct {
	concurrency int
	debug       bool
	dbName      string
	sleep       time.Duration
	mode        string
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

var cfg config

func init() {
	flag.StringVar(&cfg.dsn.host, "h", "localhost", "MySQL host")
	flag.StringVar(&cfg.dsn.port, "P", "3306", "MySQL server port")
	flag.StringVar(&cfg.dsn.user, "u", "root", "MySQL user")
	flag.StringVar(&cfg.dsn.passwd, "p", "", "MySQL password")
	flag.StringVar(&cfg.dsn.parseTime, "T", "true", "MySQL parseTime(true|false)")
	flag.IntVar(&cfg.concurrency, "c", 50, "Number of Goroutione")
	flag.BoolVar(&cfg.debug, "debug", false, "show debug level log")
	flag.StringVar(&cfg.dbName, "d", "", "MySQL database name")
	flag.DurationVar(&cfg.sleep, "sleep", 1*time.Second, "Sleep time, support time duration [s|m|h]")
	flag.StringVar(&cfg.mode, "mode", "HealthCheck", "[HealthCheck|Read|Write|RW]")
}

func main() {
	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version: %s", version)
		os.Exit(0)
	}

	log.Info().Msg("Starting program...")
	dsn := getDSN(cfg)
	conn, err := initDB(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("DB connection pool initialization failed")
	}
	defer conn.Close()
	log.Info().Msg("DB connection pool initialized")

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

	if cfg.dbName == "" && cfg.mode != "HealthCheck" {
		app.log.Warn().Str("mode", cfg.mode).Msg("Not supported when db name is not provided, will change to HealthCheck mode")
		cfg.mode = "HealthCheck"
	}

	if cfg.mode == "Write" || cfg.mode == "RW" {
		if err = app.db.CreateDB(cfg.dbName); err != nil {
			app.log.Fatal().Err(err).Msg("Database init failed")
		}
		if err = app.db.CreateTab(cfg.dbName); err != nil {
			app.log.Fatal().Err(err).Msg("Create test table failed")
		}
		app.log.Info().Msg("Database initialization complete")
	}

	app.log.Info().Msg("Start to send query to MySQL...")
	for {
		c1 := make(chan int, cfg.concurrency)
		c2 := make(chan string, cfg.concurrency)
		if cfg.mode == "Write" || cfg.mode == "RW" {
			go func() {
				for x := 0; x < cfg.concurrency; x++ {
					c1 <- random.Integer(100000000)
					c2 <- random.String(20)
				}
			}()
		}
		var wg sync.WaitGroup
		wg.Add(cfg.concurrency)

		var master *string
		for i := 0; i < cfg.concurrency; i++ {
			go func(i int) {
				switch cfg.mode {
				case "HealthCheck":
					master, err = app.db.GetSysVar("wsrep_node_name")
				case "Write":
					col1 := <-c1
					col2 := <-c2
					master, err = app.db.Insert(cfg.dbName, col1, col2)
				case "RW":
					col1 := <-c1
					col2 := <-c2
					master, err = app.db.Txn(cfg.dbName, col1, col2)
				default:
					app.log.Warn().Str("mode", cfg.mode).Msg("Unsupported mode, using HealthCheck")
					cfg.mode = "HealthCheck"
					master, err = app.db.GetSysVar("wsrep_node_name")
				}
				if err != nil {
					app.log.Error().Err(err).Str("mode", cfg.mode).Int("Goroutine", i).Msg("Failed")
				} else {
					if master != nil {
						app.log.Info().Str("mode", cfg.mode).Int("Goroutine", i).Str("master", *master).Msg("Query Succeed")
					} else {
						app.log.Info().Str("mode", cfg.mode).Int("Goroutine", i).Msg("Query Succeed")
					}
				}
				wg.Done()
			}(i)
		}
		wg.Wait()
		time.Sleep(cfg.sleep)
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
