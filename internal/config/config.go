package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI           string
	MongoDB            string
	ServerPort         string
	JwtSecret          string
	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseBucket     string
}

func extractEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("missing required key (%s)", key)
	}
	return value, nil
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		return Config{}, fmt.Errorf("failed to load .env -> %w", err)
	}

	mongoURI, err := extractEnv("MONGO_URI")
	if err != nil {
		return Config{}, fmt.Errorf("%w", err)
	}

	mongoDB, err := extractEnv("MONGO_DB_NAME")
	if err != nil {
		return Config{}, fmt.Errorf("%w", err)
	}

	serverPort, err := extractEnv("PORT")
	if err != nil {
		return Config{}, fmt.Errorf("%w", err)
	}

	jwtSecret, err := extractEnv("JWT_SECRET")
	if err != nil {
		return Config{}, fmt.Errorf("%w", err)
	}

	supabaseURL, err := extractEnv("SUPABASE_URL")
	if err != nil {
		return Config{}, fmt.Errorf("%w", err)
	}

	supabaseServiceKey, err := extractEnv("SUPABASE_SERVICE_KEY")
	if err != nil {
		return Config{}, fmt.Errorf("%w", err)
	}

	supabaseBucket, err := extractEnv("SUPABASE_BUCKET")
	if err != nil {
		return Config{}, fmt.Errorf("%w", err)
	}

	return Config{
		MongoURI:           mongoURI,
		MongoDB:            mongoDB,
		ServerPort:         serverPort,
		JwtSecret:          jwtSecret,
		SupabaseURL:        supabaseURL,
		SupabaseServiceKey: supabaseServiceKey,
		SupabaseBucket:     supabaseBucket,
	}, nil
}
