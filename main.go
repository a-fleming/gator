package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"www.github.com/a-fleming/blog-aggregator/internal/config"
	"www.github.com/a-fleming/blog-aggregator/internal/database"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("gator: error: the following arguments are required: command")
		os.Exit(1)
	}
	commandName := os.Args[1]
	args := os.Args[2:]

	var cfg config.Config
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	cliState := state{
		config: &cfg,
		db:     dbQueries,
	}

	cmds := GetCommands()

	cmdToRun := command{
		name:      commandName,
		arguments: args,
	}

	err = cmds.run(&cliState, cmdToRun)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
