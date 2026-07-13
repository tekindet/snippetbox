package pkg

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr        string
	StaticDir   string
	SecretKey   string
	DatabaseURI string
}

func getVarFromEnv(envName string) (string, error) {
	env, exists := os.LookupEnv(envName)
	if !exists {
		return "", fmt.Errorf("config: env variable %s is not set", envName)
	}
	return env, nil
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	addr, err := getVarFromEnv("APP_ADDR")
	if err != nil {
		return nil, err
	}
	staticDir, err := getVarFromEnv("STATIC_DIR")
	if err != nil {
		return nil, err
	}
	secretKey, err := getVarFromEnv("SECRET_KEY")
	if err != nil {
		return nil, err
	}
	databaseURI, err := getVarFromEnv("DATABASE_URI")
	if err != nil {
		return nil, err
	}

	return &Config{
		Addr:        addr,
		StaticDir:   staticDir,
		SecretKey:   secretKey,
		DatabaseURI: databaseURI,
	}, nil
}
