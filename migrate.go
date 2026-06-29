//go:build ignore
// +build ignore

package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	// Ganti sesuai kebutuhan Anda
	migratePath := "internal/infrastructure/database/sql"
	dbURL := "postgres://postgres:rss@j@y@2025@10.10.123.206:5432/health?sslmode=disable"

	args := append([]string{"-path", migratePath, "-database", dbURL}, os.Args[1:]...)
	cmd := exec.Command("migrate", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
