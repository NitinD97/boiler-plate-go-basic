package config

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/viper"
)

var CONFIG_PATH_MAP = map[string]string{
	"development": "./config/development.json",
	"staging":     "./config/staging.json",
	"production":  "./config/production.json",
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetBoolean(key string) bool {
	return viper.GetBool(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetStringMap(key string) map[string]string {
	return viper.GetStringMapString(key)
}

func GetSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func GetMap(key string) map[string]interface{} {
	return viper.GetStringMap(key)
}

func Init() {
	runtimeEnv := flag.String("env", "development", "runtime environment")
	envFile := flag.String("env_file", "./.env", "env file path")
	flag.Parse()

	// Fetch config file based on the runtime environment
	path, ok := CONFIG_PATH_MAP[*runtimeEnv]
	if !ok {
		panic("Invalid runtime environment")
	}

	viper.SetConfigType("json")
	reader, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("unable to read config file\n %w", err))
	}
	defer reader.Close()

	err = viper.ReadConfig(reader)
	if err != nil {
		panic(fmt.Errorf("unable to read config file\n %w", err))
	}

	// Read .env file
	reader, err = os.Open(*envFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If .env file does not exist, continue without it
			loadSystemEnv()
			return
		}
		panic(fmt.Errorf("unable to read env config file\n %w", err))
	}
	defer reader.Close()

	viper.SetConfigType("env")
	envData, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("unable to read config file data\n %w", err))
	}

	// Format env keys from .env file and merge with existing config
	envData = formatEnvKeys(envData)
	viper.MergeConfig(bytes.NewBuffer(envData))

	// Load system environment variables as well
	loadSystemEnv()
}

func loadSystemEnv() {
	// Fetch all system environment variables
	envVars := os.Environ()
	formattedEnv := make([]byte, 0)

	for _, env := range envVars {
		// Split each environment variable by "="
		splits := strings.SplitN(env, "=", 2)
		if len(splits) != 2 {
			continue
		}

		// Replace "_" with "." in the key and format it
		newKey := strings.ReplaceAll(splits[0], "_", ".")
		formattedEnv = append(formattedEnv, []byte(newKey+"="+splits[1]+"\n")...)
	}

	// Merge system environment variables into Viper configuration
	viper.MergeConfig(bytes.NewBuffer(formattedEnv))
}

func formatEnvKeys(envData []byte) []byte {
	formattedEnv := make([]byte, 0, len(envData))
	data := strings.Split(string(envData), "\n")
	for _, line := range data {
		if line == "" {
			continue
		}
		splits := strings.SplitN(line, "=", 2)
		newKKey := strings.ReplaceAll(splits[0], "~", ".")
		formattedEnv = append(formattedEnv, []byte(newKKey+"="+splits[1]+"\n")...)
	}
	return formattedEnv
}
