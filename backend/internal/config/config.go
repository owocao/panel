package config

import "os"

type Config struct {
	Port      string
	DataDir   string
	DBPath    string
	StaticDir string
	AdminUser string
	AdminPass string
}

func Load() Config {
	dataDir := getEnv("BIU_PANEL_DATA_DIR", "../data")
	return Config{
		Port:      getEnv("BIU_PANEL_PORT", "55088"),
		DataDir:   dataDir,
		DBPath:    getEnv("BIU_PANEL_DB", dataDir+"/db/biu-panel.db"),
		StaticDir: getEnv("BIU_PANEL_STATIC_DIR", "./public"),
		AdminUser: os.Getenv("BIU_PANEL_ADMIN_USER"),
		AdminPass: os.Getenv("BIU_PANEL_ADMIN_PASSWORD"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
