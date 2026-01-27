package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string, timeoutSec int) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	var feedData RSSFeed
	err = xml.Unmarshal(data, &feedData)
	if err != nil {
		return &RSSFeed{}, err
	}

	feedData.Channel.Title = html.UnescapeString(feedData.Channel.Title)
	feedData.Channel.Link = html.UnescapeString(feedData.Channel.Link)
	feedData.Channel.Description = html.UnescapeString(feedData.Channel.Description)
	for _, item := range feedData.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Link = html.UnescapeString(item.Link)
		item.Description = html.UnescapeString(item.Description)
		item.PubDate = html.UnescapeString(item.PubDate)
	}

	return &feedData, nil
}

func printFeed(feed RSSFeed) {
	fmt.Printf("RSS Feed: %s\n", feed.Channel.Title)
	fmt.Printf("Link: %s\n", feed.Channel.Link)
	fmt.Printf("Description: %s\n", feed.Channel.Description)
	for _, item := range feed.Channel.Item {
		fmt.Printf(" * Title: %s\n", item.Title)
	}
}

func scrapeFeeds(s *state, timeoutSec int) error {
	ctx := context.Background()
	feedDbInfo, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}
	rssFeed, err := fetchFeed(ctx, feedDbInfo.Url, timeoutSec)
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(ctx, feedDbInfo.ID)
	if err != nil {
		return err
	}
	printFeed(*rssFeed)
	return nil
}
