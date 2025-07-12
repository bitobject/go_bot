package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// URL для проверки состояния берется из аргументов командной строки,
	// со значением по умолчанию для локального запуска.
	url := "http://localhost:8080/health"
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Health check failed with status: %s\n", resp.Status)
		os.Exit(1)
	}

	fmt.Println("Health check passed successfully.")
	os.Exit(0)
}
