package postgres

import (
	"log"
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
	cmd.Env = append(cmd.Env,
		"GOOSE_DRIVER=postgres",
		"GOOSE_DBSTRING=host=localhost user=postgres password=postgres database="+dbname+" sslmode=disable",
	)

	// Выполняем команду и захватываем вывод
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v\nOutput: %s", err, string(output))
	}

	log.Printf("Migrations applied successfully:\n%s", string(output))
}
