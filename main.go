package main

import (
	"database/sql"
	"fmt"
	"github.com/Vvandereer/Observer/internal/config"
	"github.com/Vvandereer/Observer/internal/database"
	_ "github.com/lib/pq"
	"os"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

func main() {
	fmt.Println("Hello from Observer!")

	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	s := state{
		cfg: &cfg,
		db:  dbQueries,
	}

	cmds := commands{
		list: make(map[string]func(*state, command) error),
	}

	CommandsInit(cmds)
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments were provided")
		os.Exit(1)
	}
	cmd := command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}
	if err := cmds.run(&s, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
