package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite")

var DB *sql.DB

func Connect() {
	db, err := sql.Open("sqlite", "./mangahub.db")
	if err != nil {
		log.Fatal("Cannot open database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	DB = db
	log.Println("Connected to SQLite successfully")
}

func Migrate() {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);
	`

	mangaTable := `
	CREATE TABLE IF NOT EXISTS manga (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT,
		genres TEXT,
		status TEXT,
		total_chapters INTEGER,
		description TEXT
	);
	`

	progressTable := `
	CREATE TABLE IF NOT EXISTS user_progress (
		user_id INTEGER,
		manga_id TEXT,
		current_chapter INTEGER DEFAULT 0,
		status TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, manga_id)
	);
	`

	_, err := DB.Exec(userTable)
	if err != nil {
		log.Fatal("Create users table failed:", err)
	}

	_, err = DB.Exec(mangaTable)
	if err != nil {
		log.Fatal("Create manga table failed:", err)
	}

	_, err = DB.Exec(progressTable)
	if err != nil {
		log.Fatal("Create user_progress table failed:", err)
	}

	log.Println("Database migrated successfully")
}