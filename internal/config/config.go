package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func LoadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Println("Warning: Invalid .env line:", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading .env file:", err)
	}
}
