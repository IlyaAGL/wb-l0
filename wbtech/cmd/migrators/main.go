package main

import (
	"github.com/agl/wbtech/internal/infrastructure/migrations"
	"github.com/agl/wbtech/pkg/dbconnections"
)

func main() {
	db_pg := dbconnections.InitPostgres()
	defer db_pg.Close()

	migrations.RunMigrationsPG(db_pg)
}
