package main

import (
	"github.com/jinzhu/gorm"
	tracinggorm "github.com/tsocial/tracing/gorm"
	"github.com/tsocial/vite"
	"log"
	"time"
)

var testDB *gorm.DB

// Files contain incremental migration SQL scripts to create or drop the schemas
const (
	CreateOpenCensusSchemaFile = "create_open_census_schema.sql"
	DropOpenCensusSchemaFile   = "drop_open_census_schema.sql"
)

const (
	dbServer                    = "0.0.0.0:3306"
	dbSchema                    = "open_census"
	dbUser                      = "admin"
	dbPassword                  = "password"
	dbConnectionLifetimeSeconds = 300
	dbMaxIdleConnection         = 0
	dbMaxOpenConnection         = 1
)

func createStandardDBConnection() *gorm.DB {
	config := &vite.MySQLConfig{
		Server:                    vite.EVString("DB_SERVER", dbServer),
		Schema:                    vite.EVString("DB_SCHEMA", dbSchema),
		User:                      vite.EVString("DB_USER", dbUser),
		Password:                  vite.EVString("DB_PASSWORD", dbPassword),
		ConnectionLifetimeSeconds: vite.EVInt("DB_CONNECTION_LIFETIME_SECONDS", dbConnectionLifetimeSeconds),
		MaxIdleConnections:        vite.EVInt("DB_MAX_IDLE_CONNECTIONS", dbMaxIdleConnection),
		MaxOpenConnections:        vite.EVInt("DB_MAX_OPEN_CONNECTIONS", dbMaxOpenConnection),
	}
	var err error
	testDB, err = vite.ConnectORM(config)
	if err != nil {
		panic(err)
	}
	return testDB
}

func MigrateDB() {
	testDB = createStandardDBConnection()
	err := vite.MigrateORM(testDB, CreateOpenCensusSchemaFile)
	if err != nil {
		panic(err)
	}

	tracinggorm.RegisterGormCallbacks(testDB)
}

type Product struct {
	ID        int64     `json:"id" gorm:"id"`
	Name      string    `json:"name" gorm:"name"`
	Price     int       `json:"price" gorm:"price"`
	CreatedAt time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"updated_at"`
}

func GetFirstProduct() (*Product, error) {
	r := &Product{}
	if err := testDB.First(r, 1).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		log.Println(vite.MarkError, err)
		return nil, err
	}

	return r, nil
}

func GetFirstProductWithContext(db *gorm.DB) (*Product, error) {
	r := &Product{}
	if err := db.First(r, 1).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		log.Println(vite.MarkError, err)
		return nil, err
	}

	return r, nil
}

func GetAllProductDates(db *gorm.DB) error {
	r := &Product{}
	rows, err := db.Table("products").Where(r).Select("date(created_at) as date").Rows()
	defer func() {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		log.Println(vite.MarkError, err)
		return err
	}
	return nil
}
