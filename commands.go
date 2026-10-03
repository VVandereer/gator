package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/Vvandereer/Observer/internal/database"
	"github.com/google/uuid"
	"html"
	"io"
	"net/http"
	"os"
	"time"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	list map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) error {
	if _, exist := c.list[name]; exist {
		return fmt.Errorf("Command %s already exists", name)
	}
	c.list[name] = f
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	if com, exist := c.list[cmd.Name]; exist {
		return com(s, cmd)
	}
	return fmt.Errorf("Command doesn't exists")
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("required username")
	}
	user, err := (*(s.db)).CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	})
	if err != nil {
		return err
	}
	(*s.cfg).SetUser(cmd.Args[0])
	fmt.Printf("User was created: ", user)
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("required username")
	}

	user, err := (*s.db).GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("unknown username")
	}
	(*s.cfg).SetUser(user.Name)
	fmt.Printf("Username has been set like %s\n", (*s.cfg).CurrentUsername)
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := (*s.db).GetUsers(context.Background())
	if err != nil {
		return err
	}
	if len(users) < 1 {
		return fmt.Errorf("No users")
	}
	for _, name := range users {
		fmt.Printf(" * %s", name)
		if (*s.cfg).CurrentUsername == name {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return nil
}

func handleReset(s *state, cmd command) error {
	err := (*s.db).Reset(context.Background())
	if err == nil {
		fmt.Println("Successful reseting table")
	}
	return err
}

func CommandsInit(cmds commands) {
	comt := map[string]func(s *state, cmd command) error{
		"login":     handlerLogin,
		"register":  handlerRegister,
		"reset":     handleReset,
		"users":     handlerGetUsers,
		"agg":       handlerAgg,
		"addfeed":   middlewareLoggedIn(handlerAddFeed),
		"feeds":     handlerFeeds,
		"follow":    middlewareLoggedIn(handlerFollow),
		"following": middlewareLoggedIn(handlerFollowing),
		"unfollow":  middlewareLoggedIn(handlerUnfollow),
	}
	for name, handler := range comt {
		if err := cmds.register(name, handler); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Atom    string   `xml:"atom,attr"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description   string `xml:"description"`
		Generator     string `xml:"generator"`
		Language      string `xml:"language"`
		LastBuildDate string `xml:"lastBuildDate"`
		Item          []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Guid        string `xml:"guid"`
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var RSS RSSFeed
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSS, err
	}
	req.Header.Set("User-Agent", "gator")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSS, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSS, err
	}
	data = []byte(html.UnescapeString(html.UnescapeString(string(data))))
	if err := xml.Unmarshal(data, &RSS); err != nil {
		return &RSS, err
	}
	return &RSS, nil
}

func handlerAgg(s *state, cmd command) error {
	feed, err := (*s.db).GetNextFeedToFetch(ctx.Background())
	RSS, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}
	fmt.Println(*RSS)
	return nil
}

func scrapeFeeds(s *state, cmd command) error {
	feed, err := (*s.db).GetNextFeedToFetch(ctx.Background())
	if err != nil {
		return err
	}
	err = (*s.db).MarkFeedFetched(ctx.Background(), feed.ID)
	if err != nil {
		return err
	}
	RSS, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}
	fmt.Println(*RSS)
	for item := range feed {
		fmt.Println(item)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("required feed name and url")
	}

	feed, err := (*s.db).CreateFeed(context.Background(), database.CreateFeedParams{
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}
	FollowParams := database.CreateFeedFollowParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = (*s.db).CreateFeedFollow(context.Background(), FollowParams)
	if err != nil {
		return err
	}
	fmt.Printf("Feed was created: ", feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := (*s.db).GetFeeds(context.Background())
	if err != nil {
		return err
	}
	if len(feeds) < 1 {
		return fmt.Errorf("No feeds")
	}
	for _, feed := range feeds {
		var username string
		user, err := (*s.db).GetUserById(context.Background(), feed.UserID)
		if err != nil {
			username = "unknown user"
		} else {
			username = user.Name
		}
		fmt.Printf(" * %s | %s | added by %s\n", feed.Name, feed.Url, username)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("url required")
	}
	url := cmd.Args[0]
	feed, err := (*s.db).GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}
	FollowParams := database.CreateFeedFollowParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	_, err = (*s.db).CreateFeedFollow(context.Background(), FollowParams)
	if err != nil {
		return err
	}
	fmt.Printf(" %s %s\n", feed.Name, user.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := (*s.db).GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	for _, name := range follows {
		fmt.Println(name)
	}
	return nil

}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("url required")
	}
	url := cmd.Args[0]
	feed, err := (*s.db).GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}
	if err := (*s.db).DelFeedFollow(context.Background(), database.DelFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}); err != nil {
		return err
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, cmd command) error {
		usr, err := (*s.db).GetUser(context.Background(), (*s.cfg).CurrentUsername)
		if err != nil {
			return err
		}
		user := database.User{
			ID:        usr.ID,
			CreatedAt: usr.CreatedAt,
			UpdatedAt: usr.UpdatedAt,
			Name:      usr.Name,
		}
		return handler(s, cmd, user)
	}
}
