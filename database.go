package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func initDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	createTables()
}
func createTables() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		telegram_token TEXT,
		telegram_chat_id INTEGER,
		created_at TIMESTAMP DEFAULT NOW()
	)
`)
	if err != nil {
		log.Fatal("Error creating users table:", err)
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS services (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            address TEXT NOT NULL,
            phone TEXT NOT NULL,
            logo_path TEXT,
            owner_id INTEGER NOT NULL,
            FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
			latitude DECIMAL(10, 7),
			longitude DECIMAL(10, 7),
			approved BOOLEAN DEFAULT FALSE,
			description TEXT DEFAULT 'Отсутствует',
			working_hours JSONB,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		log.Fatal("Error creating services table:", err)
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_services_owner_id ON services(owner_id)")
	if err != nil {
		log.Fatal("Error creating index:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS admin_logs (
			id SERIAL PRIMARY KEY,
			admin_id INT REFERENCES users(id),
			action TEXT,
			created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		log.Fatal("Error creating admin_logs table:", err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS bookings (
		id SERIAL PRIMARY KEY,
		partner_id INTEGER,
		user_id INTEGER,
		booking_date TEXT,
		booking_time TEXT,
		status TEXT,
		FOREIGN KEY (partner_id) REFERENCES services(id) ON DELETE CASCADE
	);`)
	if err != nil {
		log.Fatal("Error creating bookings table:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS offerings (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		)
	`)
	if err != nil {
		log.Fatal("Error creating offerings table:", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS partner_offerings (
			id SERIAL PRIMARY KEY,
			partner_id INTEGER NOT NULL,
			offering_id INTEGER NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			image_url TEXT,
			FOREIGN KEY (partner_id) REFERENCES services(id) ON DELETE CASCADE,
			FOREIGN KEY (offering_id) REFERENCES offerings(id)
		)
	`)
	if err != nil {
		log.Fatal("Error creating partner_offerings table:", err)
	}
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS announcements (
        id SERIAL PRIMARY KEY,
        partner_id INTEGER NOT NULL,
        title TEXT NOT NULL,
        text TEXT NOT NULL,
        image_url TEXT,
        FOREIGN KEY (partner_id) REFERENCES services(id) ON DELETE CASCADE
    )
`)
	if err != nil {
		log.Fatal("Error creating announcements table:", err)
	}
}
