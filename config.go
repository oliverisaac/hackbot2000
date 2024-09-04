package main

import (
	"fmt"
	sc "github.com/oliverisaac/snakeCharm"
	"github.com/spf13/viper"
)

func doConfig() *viper.Viper {
	cfg, err := sc.BuildConfig(nil, []*sc.ConfigEntry{
		{
			Type:     sc.IntType,
			Name:     "port",
			Help:     "Port on which to listen for requests",
			Required: true,
			Default:  80,
		},
		{
			Type: sc.ParentType,
			Name: "slack",
			Children: []*sc.ConfigEntry{
				{
					Type:     sc.StringType,
					Name:     "token",
					Help:     "Token to use when dealing with slack",
					Required: false,
					Default:  "",
				},
				{
					Type:     sc.StringType,
					Name:     "clientID",
					Help:     "ClientID to use with slack",
					Required: false,
					Default:  "",
				},
				{
					Type:     sc.StringType,
					Name:     "clientSecret",
					Help:     "ClientSecret to use with slack",
					Required: false,
					Default:  "",
				},
				{
					Type:     sc.StringType,
					Name:     "signingSecret",
					Help:     "SigningSecret to use with slack",
					Required: false,
					Default:  "",
				},
			},
		},
		{
			Type: sc.ParentType,
			Name: "db",
			Children: []*sc.ConfigEntry{
				{
					Type:     sc.IntType,
					Name:     "port",
					Help:     "Port on which to connect to the db",
					Required: true,
					Default:  3306,
				},
				{
					Type:     sc.StringType,
					Name:     "host",
					Help:     "Hostname to connect to in the db",
					Required: true,
					Default:  "localhost",
				},
				{
					Type:     sc.StringType,
					Name:     "username",
					Help:     "Username to use when connecting to the DB",
					Required: true,
					Default:  "",
				},
				{
					Type:     sc.StringType,
					Name:     "password",
					Help:     "Password to use when connecting to the DB",
					Required: true,
					Default:  "",
				},
				{
					Type:     sc.StringType,
					Name:     "name",
					Help:     "Name of database to use",
					Required: true,
					Default:  "",
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("port: %d\n", cfg.GetInt("port"))
	fmt.Printf("db.port: %d\n", cfg.GetInt("db.port"))
	fmt.Printf("db.host: %s\n", cfg.GetString("db.host"))
	fmt.Printf("db.username: %s\n", cfg.GetString("db.username"))
	fmt.Printf("db.password: %s\n", cfg.GetString("db.password"))
	fmt.Printf("db.name: %s\n", cfg.GetString("db.name"))
	fmt.Printf("slack.token: %s\n", cfg.GetString("slack.token"))

	return cfg
}
