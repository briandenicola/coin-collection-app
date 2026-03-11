package config

import "os"

type Config struct {
	DBPath    string
	JWTSecret string
	OllamaURL string
	Port      string
	UploadDir string
}

func Load() *Config {
	return &Config{
		DBPath:    getEnv("DB_PATH", "./ancientcoins.db"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-key-change-in-production-min32chars"),
		OllamaURL: getEnv("OLLAMA_URL", "http://localhost:11434"),
		Port:      getEnv("PORT", "8080"),
		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
