package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	App struct {
		Host            string
		Port            string
		EnableProfiling bool
	}

	Db struct {
		Path string
	}

	Index struct {
		Path string
	}
}

func NewConfig() (*config, error) {
	configPath := flag.String("config", "", "full path to .env config file")

	appHost := flag.String("host", "0.0.0.0", "host where app will run")
	appPort := flag.String("port", "8000", "port where app will run")
	appEnableProfiling := flag.Bool("profiling", false, "enable gin pprof profiling endpoints")

	dbFilePath := flag.String("dbpath", "db.sqlite3", "full path to .sqlite3 db file")

	indexFilePath := flag.String("indexpath", "index.bleve", "full path to .bleve index file")

	flag.Parse()

	if *configPath != "" {
		err := godotenv.Load(*configPath)
		if err != nil {
			panic(err)
		}

		if env, ok := os.LookupEnv("APP_HOST"); ok {
			*appHost = env
		}

		if env, ok := os.LookupEnv("APP_PORT"); ok {
			*appPort = env
		}

		if env, ok := os.LookupEnv("APP_ENABLE_PROFILING"); ok {
			*appEnableProfiling, err = strconv.ParseBool(env)
			if err != nil {
				panic(err)
			}
		}

		if env, ok := os.LookupEnv("DB_FILE_PATH"); ok {
			*dbFilePath = env
		}

		if env, ok := os.LookupEnv("INDEX_FILE_PATH"); ok {
			*indexFilePath = env
		}
	}

	return &config{
		App: struct {
			Host            string
			Port            string
			EnableProfiling bool
		}{
			*appHost,
			*appPort,
			*appEnableProfiling,
		},
		Db:    struct{ Path string }{*dbFilePath},
		Index: struct{ Path string }{*indexFilePath},
	}, nil
}
