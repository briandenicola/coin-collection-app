package models

type AppSetting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}
