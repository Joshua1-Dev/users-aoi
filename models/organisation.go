package models

import (
	"gorm.io/gorm"
)

type Organisation struct {
	gorm.Model
	OrgID       string `gorm:"unique" json:"orgId"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	
}
