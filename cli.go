package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/a-fleming/blog-aggregator/internal/config"
	"github.com/a-fleming/blog-aggregator/internal/database"
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
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("agg", handlerAggregate)
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("login", handlerLogin)
	cmds.register("logout", handlerLogout)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("users", handlerUsers)
	return cmds
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.config.CurrentUserID == "" {
			return fmt.Errorf("you must be logged in to run '%s'", cmd.name)
		}
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("gator addfeed: error: the following arguments are required: name url")
	}
	feedName := cmd.arguments[0]
	feedURL := cmd.arguments[1]

	createFeedParams := database.CreateFeedParams{
		Name:   feedName,
		Url:    feedURL,
		UserID: user.ID,
	}

	ctx := context.Background()
	feedInfo, err := s.db.CreateFeed(ctx, createFeedParams)
	if err != nil {
		return err
	}
	fmt.Println("successfully added RSS feed")
	fmt.Printf("id: %+v\n", feedInfo.ID)
	fmt.Printf("name: %s\n", feedName)
	fmt.Printf("url: %s\n", feedURL)
	fmt.Printf("createAt: %+v\n", feedInfo.CreatedAt)
	fmt.Printf("updatedAt: %+v\n", feedInfo.UpdatedAt)
	fmt.Printf("user_id: %+v\n", feedInfo.ID)

	createFeedFollowParams := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedInfo.ID,
	}
	followedFeed, err := s.db.CreateFeedFollow(ctx, createFeedFollowParams)
	if err != nil {
		return err
	}
	fmt.Printf("'%s' has followed '%s'\n", followedFeed.UserName, followedFeed.FeedName)
	return nil
}

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("gator agg: error: the following argument is required: time_between_requests")
	}
	timeStr := cmd.arguments[0]
	timeBetweenRequests, err := time.ParseDuration(timeStr)
	if err != nil {
		return err
	}
	fmt.Printf("collecting feeds every %s\n", timeStr)
	ticker := time.NewTicker(timeBetweenRequests)
	ctx := context.Background()
	for ; ; <-ticker.C {
		httpTimeoutSec := 15
		err = scrapeFeeds(ctx, s, httpTimeoutSec)
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println()
		fmt.Println("----------------------------------------------------------")
		fmt.Println()
		fmt.Println()
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2)
	if len(cmd.arguments) > 0 {
		parsedLimit, err := strconv.ParseInt(cmd.arguments[0], 10, 32)
		if err != nil {
			return err
		}
		limit = int32(parsedLimit)
	}
	params := database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: limit,
	}
	ctx := context.Background()
	posts, err := s.db.GetPostsForUser(ctx, params)
	if err != nil {
		return err
	}
	for idx, post := range posts {
		descriptionStr := ""
		if post.Description.Valid {
			descriptionStr = post.Description.String
		}

		fmt.Printf("%d. Title: %s\n", idx+1, post.Title)
		fmt.Printf("-- Link: %s\n", post.Url)
		fmt.Printf("-- Date: %s\n", post.PublishedAt)
		fmt.Printf("-- Description: %s\n", descriptionStr)
		fmt.Println()
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		user, err := s.db.GetUserById(ctx, feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("* %s\n", feed.Name)
		fmt.Printf("--- url: %s\n", feed.Url)
		fmt.Printf("--- added by: %s\n", user.Name)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("gator follow: error: the following argument is required: url")
	}
	feedURL := cmd.arguments[0]

	ctx := context.Background()
	feedInfo, err := s.db.GetFeedByUrl(ctx, feedURL)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return fmt.Errorf("feed not found at '%s'", feedURL)
		}
		return err
	}

	params := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedInfo.ID,
	}
	followedFeed, err := s.db.CreateFeedFollow(ctx, params)
	if err != nil {
		return err
	}
	fmt.Printf("'%s' has followed '%s'\n", followedFeed.UserName, followedFeed.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	feedFollows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}
	if len(feedFollows) == 0 {
		fmt.Printf("'%s' is not following any feeds\n", s.config.CurrentUserName)
		return nil
	}

	fmt.Printf("'%s' is following:\n", s.config.CurrentUserName)
	for _, feedFollow := range feedFollows {
		fmt.Printf("* '%s' (%s)\n", feedFollow.FeedName, feedFollow.FeedUrl)
	}
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("gator login: error: the following argument is required: username")
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

func handlerLogout(s *state, cmd command) error {
	loggedOutUserName := s.config.CurrentUserName
	err := s.config.SetUser("", "")
	if err != nil {
		return err
	}
	fmt.Printf("'%s' was logged out\n", loggedOutUserName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("gator register: error: the following argument is required: username")
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
	fmt.Printf("user '%s' was created\n", userName)
	fmt.Printf("%+v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.Reset(ctx)
	if err != nil {
		return err
	}
	err = handlerLogout(s, cmd)
	if err != nil {
		return err
	}
	fmt.Println("successfully reset 'users' table")
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("gator unfollow: error: the following argument is required: url")
	}
	feedURL := cmd.arguments[0]
	ctx := context.Background()
	feed, err := s.db.GetFeedByUrl(ctx, feedURL)
	if err != nil {
		return err
	}
	params := database.RemoveFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.RemoveFeedFollow(ctx, params)
	if err != nil {
		return err
	}
	fmt.Printf("successfully unfollowed feed '%s'\n", feed.Name)
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
		return fmt.Errorf("unknown command '%s'", cmd.name)
	}
	return cmdFunc(s, cmd)
}
