package main

type Config struct {
	Listen    string `dhall:"listen" json:"listen"`
	Auth      Auth   `dhall:"auth" json:"auth"`
	Executors int    `dhall:"executors" json:"executors"`
	Menus     []Menu `dhall:"menus" json:"menus"`
}

type Auth struct {
	Username string `dhall:"username" json:"username"`
	Password string `dhall:"password" json:"password"`
}

type Menu struct {
	Link      *LinkMenu      `dhall:"link" json:"link"`
	Selection *SelectionMenu `dhall:"selection" json:"selection"`
}

type LinkMenu struct {
	Patterns []Pattern `dhall:"patterns" json:"patterns"`
	Actions  []Action  `dhall:"actions" json:"actions"`
}

type SelectionMenu struct {
	Regexes []string `dhall:"regexes" json:"regexes"`
	Actions []Action `dhall:"actions" json:"actions"`
}

type Pattern struct {
	URL   string `dhall:"url" json:"url"`
	Regex string `dhall:"regex" json:"regex"`
}

type Action struct {
	Name   string `dhall:"name" json:"name"`
	Script string `dhall:"script" json:"script"`
}
