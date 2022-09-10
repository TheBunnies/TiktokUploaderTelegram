package db

import (
	"gorm.io/gorm"
	"time"
)

type Driver struct {
	db *gorm.DB
}

type Log struct {
	Id           int `gorm:"primaryKey"`
	Message      string
	Type         string
	CreationTime time.Time
}

type User struct {
	Id        int64 `gorm:"primaryKey"`
	FirstName string
	LastName  string
	Username  string
}
