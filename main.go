package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dreamspawn/ribbon-server/api"
	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/connect"
	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/log"
	"github.com/dreamspawn/ribbon-server/render"
	"github.com/dreamspawn/ribbon-server/server"
	"github.com/dreamspawn/ribbon-server/translations"
)

func main() {
	file := os.Args[0]
	dir := "."
	index := strings.LastIndex(file, "/")
	if index != -1 {
		dir = file[:index]
	}

	config_path := dir + "/ribbon-server.conf"
	err := config.LoadConfig(config_path)
	if err != nil {
		fmt.Println("Could not open config file: " + config_path)
		fmt.Println(err.Error())
		fmt.Println("Do you want to create a new configuration file? (Y/n)")

		var input string
		fmt.Scanln(&input)
		if input == "Y" || input == "y" || input == "" {
			config.CreateConfig(config_path)
		} else {
			fmt.Println("Can't continue without a configuration file, closing server")
			os.Exit(1)
		}
	}

	log.Init()

	server.ConfigReady()
	connect.ConfigReady()
	render.ConfigReady()
	api.ConfigReady()
	translations.ConfigReady()
	translations.LoadAll()

	// Config test
	fmt.Printf("Resouce directory: %s\n", config.Get("resource_dir"))

	database.Connect()
	database.Update()

	var ribbon_server = server.Server{}
	ribbon_server.Start(config.Get("socket_path"))

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	<-channel

	ribbon_server.Stop()
	fmt.Println("")
}
