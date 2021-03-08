// This is the main class.
// Where you will extract the inputs asked on the config.json file and call the formula's method(s).

package main

import (
	"os"
	"postgres/exec/command/pkg/formula"
)

func main() {
	commandType := os.Getenv("COMMAND_TYPE")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")
	dbSsl := os.Getenv("DB_SSL")

	formula.Formula{
		CommandType: commandType,
		DBHost:      dbHost,
		DBName:      dbName,
		DBUsername:  dbUsername,
		DBPassword:  dbPassword,
		DBPort:      dbPort,
		DBSsl:       dbSsl,
	}.Run(os.Stdout)
}