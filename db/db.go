package db

import (
	"github.com/TheBunnies/TiktokUploaderTelegram/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

var (
	DRIVER = Driver{}
)

func Setup() {
	db, err := gorm.Open(postgres.Open(config.ConStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Log{})

	DRIVER.db = db
}

func (driver Driver) CreateUser(id int64, firstName string, lastName string, username string) error {
	user := User{
		Id:        id,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
	}
	if result := driver.db.Create(&user); result.Error != nil {
		return result.Error
	}
	//cache.Cache.Set([]byte(fmt.Sprint(id)), []byte("exists"), 600)
	return nil
}

func (driver Driver) GetUser(id int64) (*User, error) {
	var user User
	if result := driver.db.First(&user, id); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (driver Driver) UpdateUser(oldUser User, newUser User) error {
	if result := driver.db.Model(&oldUser).Updates(newUser); result.Error != nil {
		return result.Error
	}
	return nil
}

func (driver Driver) IsUserExists(id int64) (bool, error) {
	var exists bool
	/*_, err := cache.Cache.Get([]byte(fmt.Sprint(id)))
	if err == nil {
		return true, nil
	}*/

	err := driver.db.Model(&User{}).
		Select("count(*) > 0").
		Where("id = ?", id).
		Find(&exists).
		Error

	//cache.Cache.Set([]byte(fmt.Sprint(id)), []byte("exists"), 600)
	return exists, err
}

func (driver Driver) LogInformation(message ...string) {
	logRecord := Log{
		Message:      strings.Join(message, " "),
		Type:         "info",
		CreationTime: time.Now(),
	}
	if result := driver.db.Create(&logRecord); result.Error != nil {
		log.Println("Error while sinking logs into the database", result.Error)
		log.Println(strings.Join(message, " "))
	}
}

func (driver Driver) LogError(message ...string) {
	logRecord := Log{
		Message:      strings.Join(message, " "),
		Type:         "error",
		CreationTime: time.Now(),
	}
	if result := driver.db.Create(&logRecord); result.Error != nil {
		log.Println("Error while sinking logs into the database", result.Error)
		log.Println(strings.Join(message, " "))
	}
}
