package services

import (
	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
)

const (
	SettingOllamaURL     = "OllamaURL"
	SettingOllamaModel   = "OllamaModel"
	SettingAIPrompt      = "AiAnalysisPrompt"
	SettingOllamaTimeout = "OllamaTimeout"
	SettingLogLevel      = "LogLevel"
)

var settingDefaults = map[string]string{
	SettingOllamaURL:     "http://localhost:11434",
	SettingOllamaModel:   "llava",
	SettingAIPrompt:      "",
	SettingOllamaTimeout: "300",
	SettingLogLevel:      "info",
}

func GetSetting(key string) string {
	var setting models.AppSetting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		if def, ok := settingDefaults[key]; ok {
			return def
		}
		return ""
	}
	return setting.Value
}

func SetSetting(key, value string) error {
	var setting models.AppSetting
	result := database.DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		setting = models.AppSetting{Key: key, Value: value}
		return database.DB.Create(&setting).Error
	}
	setting.Value = value
	return database.DB.Save(&setting).Error
}

func GetAllSettings() map[string]string {
	result := make(map[string]string)
	for k, v := range settingDefaults {
		result[k] = v
	}

	var settings []models.AppSetting
	database.DB.Find(&settings)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result
}
