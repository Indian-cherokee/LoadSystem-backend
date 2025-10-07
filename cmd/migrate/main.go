package main

import (
	"web/internal/app/ds"
	"web/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		panic("failed to connect database")
	}
	mig := db.Migrator()
	if !mig.HasTable(&ds.Loads{}) {
		if err = mig.CreateTable(&ds.Loads{}); err != nil {
			panic("cant create table loads")
		}
	}
	if !mig.HasTable(&ds.LoadSession{}) {
		if err = mig.CreateTable(&ds.LoadSession{}); err != nil {
			panic("cant create table load_sessions")
		}
	}
	if !mig.HasTable(&ds.Users{}) {
		if err = mig.CreateTable(&ds.Users{}); err != nil {
			panic("cant create table users")
		}
	}
	if !mig.HasTable(&ds.LoadToCalculation{}) {
		if err = mig.CreateTable(&ds.LoadToCalculation{}); err != nil {
			panic("cant create table load_to_calculations")
		}
	}

}
