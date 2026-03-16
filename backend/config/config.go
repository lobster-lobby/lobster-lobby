package config

import "os"

type Config struct {
	Port            string
	MongoDBURI      string
	MongoDBDatabase string
	JWTSecret       string
	CORSOrigins     string
	Env             string
	MeilisearchURL  string
	MeilisearchKey  string
	RebuildIndex    bool
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8090"),
		MongoDBURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase: getEnv("MONGODB_DATABASE", "lobster_lobby"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		CORSOrigins:     getEnv("CORS_ORIGINS", "http://localhost:5173"),
		Env:             getEnv("ENV", "development"),
		MeilisearchURL:  getEnv("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchKey:  getEnv("MEILISEARCH_KEY", ""),
		RebuildIndex:    os.Getenv("REBUILD_SEARCH_INDEX") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
