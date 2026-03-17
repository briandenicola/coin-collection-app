package config

import "os"

type Config struct {
	DBPath      string
	JWTSecret   string
	Port        string
	UploadDir   string
	WebAuthnID  string
	WebAuthnOrigin string
}

func Load() *Config {
	return &Config{
		DBPath:         getEnv("DB_PATH", "./ancientcoins.db"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-key-change-in-production-min32chars"),
		Port:           getEnv("PORT", "8080"),
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		WebAuthnID:     getEnv("WEBAUTHN_RP_ID", "localhost"),
		WebAuthnOrigin: getEnv("WEBAUTHN_ORIGIN", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
