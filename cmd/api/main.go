package main

import (
	"log"

	"com.github/jrovieri/golang/social/internal/db"
	"com.github/jrovieri/golang/social/internal/env"
	"com.github/jrovieri/golang/social/internal/store"
)

const version = "0.0.1"

//	@title			GopherSocial API
//	@description	API for GoSocial, a social network for gophers
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	cfg := &config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			url:          env.GetString("DB_URL", ""),
			maxOpenConns: env.GetInt("DB_MAX_OPENS_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}

	db, err := db.New(
		cfg.db.url,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("database connection established")

	appStore := store.NewStorage(db)

	app := &application{
		config: *cfg,
		store:  appStore,
	}

	log.Fatal(app.run(app.mount()))
}
