// Package db owns the GORM connection and schema migration.
package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kennyg37/tasker/web/internal/task"
)

// Open connects to Postgres using the given DSN.
func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Migrate creates or updates the tasks table to match the model.
func Migrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(&task.Task{})
}
