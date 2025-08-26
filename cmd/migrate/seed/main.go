package main

import (
	"log"

	"com.github/jrovieri/golang/social/internal/db"
	"com.github/jrovieri/golang/social/internal/env"
	"com.github/jrovieri/golang/social/internal/store"
)

func main() {
	addr := env.GetString("DB_URL", "")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store, conn)
}
