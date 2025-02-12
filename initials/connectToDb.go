package initials

import (
	"log"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//for postgrace
// var DB *gorm.DB

// func Db() {
// 	var err error
// 	dsn := os.Getenv("DATABASE")
// 	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

//		if err != nil {
//			panic("failed to connect to database")
//		}
//	}
var (
	DB   *gorm.DB
	once sync.Once
)

// Initialize the database (runs only once)
func Db() {
	once.Do(func() {
		var err error
		DB, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
		if err != nil {
			log.Fatal("❌ Failed to connect to SQLite database:", err)
		}
		log.Println("✅ Connected to SQLite database successfully!")
	})
}

// GetDB returns the initialized database instance
func GetDB() *gorm.DB {
	if DB == nil {
		Db()
	}
	return DB
}
