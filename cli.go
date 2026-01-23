package main

import (
	"context"
	"fmt"

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
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("agg", handlerAggregate)
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	return cmds
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("gator addFeed: error: the following arguments are required: name url")
	}
	feedName := cmd.arguments[0]
	feedURL := cmd.arguments[1]

	uid, err := uuid.Parse(s.config.CurrentUserID)
	params := database.CreateFeedParams{
		Name:   feedName,
		Url:    feedURL,
		UserID: uid,
	}

	ctx := context.Background()
	feedInfo, err := s.db.CreateFeed(ctx, params)
	if err != nil {
		return err
	}
	fmt.Println("Successfully added RSS feed")
	fmt.Printf("id: %+v\n", feedInfo.ID)
	fmt.Printf("name: %s\n", feedName)
	fmt.Printf("url: %s\n", feedURL)
	fmt.Printf("createAt: %+v\n", feedInfo.CreatedAt)
	fmt.Printf("updatedAt: %+v\n", feedInfo.UpdatedAt)
	fmt.Printf("user_id: %+v\n", feedInfo.ID)
	return nil
}

func handlerAggregate(s *state, cmd command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	ctx := context.Background()
	feed, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}
	fmt.Printf("feed: %+v\n", feed)
	return nil
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
	err = s.config.SetUser(userName, user.ID.String())
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
	user, err := s.db.CreateUser(ctx, userName)
	if err != nil {
		return err
	}

	err = s.config.SetUser(userName, user.ID.String())
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

func handlerUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
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
