package models

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID    string `gorm:"unique" json:"userId"`
	FirstName string `json:"firstName" `
	LastName  string `json:"lastName"`
	Email     string `gorm:"unique" json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	Phone     string `json:"phone"`
}




func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	u.Password = hashPassword(u.Password)
	return
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}
