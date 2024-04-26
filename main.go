package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/dreamspawn/ribbon-server/config"
	"github.com/dreamspawn/ribbon-server/server"
)

func main() {
	file := os.Args[0]
	dir := file[:strings.LastIndex(file, "/")]

	config_path := dir + "/config"
	conf, err := config.LoadConfig(config_path)
	if err != nil {
		fmt.Println("Could not open config file: " + config_path)
		fmt.Println(err.Error())
		fmt.Println("Do you want to create a new configuration file? (Y/n)")

		var input string
		fmt.Scanln(&input)
		if input == "Y" || input == "y" || input == "" {
			conf = config.CreateConfig(config_path)
		} else {
			fmt.Println("Can't continue without a configuration file, closing server")
			os.Exit(1)
		}
	}

	// Config test
	fmt.Printf("Resouce directory: %s\n", conf.Get("resource_dir"))

	var ribbon_server = server.Server{}
	ribbon_server.Start(conf.Get("socket_path"))

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

	ribbon_server.Stop()
	fmt.Println("")
}
