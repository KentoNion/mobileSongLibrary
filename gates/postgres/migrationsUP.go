package postgres

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func RunGooseMigrations(dbname string) {
	// Создаем команду для запуска goose
	cmd := exec.Command(
		"goose",
		"-dir", "./migrations",
		"up",
	)
	// Устанавливаем переменные окружения
	cmd.Env = append(os.Environ(),
		"GOOSE_DRIVER=postgres",
		fmt.Sprintf("GOOSE_DBSTRING=host=%v user=%v password=%v database=%v sslmode=%v",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			dbname,
			os.Getenv("DB_SSLMODE"),
		))

	// Выполняем команду и захватываем вывод
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v\nOutput: %s", err, string(output))
	}

	log.Printf("Migrations applied successfully:\n%s", string(output))
}
