package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	users "jaconsta/tickets_demo/users/lib"
)

func databaseConnection() *gorm.DB {
	dsn := "host=localhost user=johndow password=randompassword dbname=users port=6543 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	return db
}

func main() {
	db := databaseConnection()

	users.UserAppBuilder(db)
}
