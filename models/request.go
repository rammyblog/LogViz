package models

import (
	"gorm.io/gorm"
)

type Request struct {
	gorm.Model
	ResponseBody    string  `gorm:"type:text;" json:"response_body"`
	ResponseStatus  int     `gorm:"type:integer;" json:"code"`
	ResponseHeaders string  `gorm:"type:text;" json:"response_headers"`
	RequestBody     string  `gorm:"type:text;" json:"request_body"`
	Path            string  `gorm:"type:text;" json:"path"`
	Headers         string  `gorm:"type:text;" json:"headers"`
	Method          string  `gorm:"type:varchar(10);" json:"method"`
	Host            string  `gorm:"type:varchar(100);" json:"host"`
	Ipaddress       string  `gorm:"type:varchar(100);" json:"ip_address"`
	TimeSpent       float64 `gorm:"type:float;" json:"time_spent"`
}
