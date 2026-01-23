package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"www.github.com/a-fleming/blog-aggregator/internal/config"
	"www.github.com/a-fleming/blog-aggregator/internal/database"
)

type state struct {
	config *config.Config
	db     *database.Queries
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	cliCommands map[string]func(*state, command) error
}

func GetCommands() commands {
	cmds := commands{
		cliCommands: map[string]func(*state, command) error{},
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	return cmds
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("gator login: error: the following arguments are required: username")
	}
	userName := cmd.arguments[0]

	ctx := context.Background()
	user, err := s.db.GetUser(ctx, userName)
	if user == (database.User{}) {
		return fmt.Errorf("username '%s' does not exist in database", userName)
	}
	if err != nil {
		return err
	}
	err = s.config.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Printf("username '%s' has been set\n", userName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("gator register: error: the following arguments are required: username")
	}
	userName := cmd.arguments[0]

	ctx := context.Background()
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	}

	user, err := s.db.CreateUser(ctx, params)
	if err != nil {
		return err
	}

	err = s.config.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Printf("User '%s' was created\n", userName)
	fmt.Printf("%+v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.Reset(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Successfully reset 'users' table")
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cliCommands[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	cmdFunc, exists := c.cliCommands[cmd.name]
	if !exists {
		return fmt.Errorf("Unknown command '%s'\n", cmd.name)
	}
	return cmdFunc(s, cmd)
}
