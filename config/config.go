package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var default_resource_dir = "/var/www/ribbon"
var default_soket = "/var/run/ribbon.sock"

type Config struct {
	values map[string]string
}

func new_config() Config {
	var config = Config{}
	config.values = make(map[string]string)
	return config
}

func (config Config) keys() []string {
	keys := make([]string, 0, len(config.values))
	for k := range config.values {
		keys = append(keys, k)
	}
	return keys
}

func (config Config) set(key, value string) {
	config.values[key] = value
}

func (config Config) Get(key string) string {
	return config.values[key]
}

func LoadConfig(path string) (Config, error) {
	config := new_config()

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	reader := bufio.NewReader(file)

	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return Config{}, err
		}

		tokens := strings.Split(string(bytes), "=")
		config.set(strings.Trim(tokens[0], " "), strings.Trim(tokens[1], " "))
	}

	return config, nil
}

func CreateConfig(path string) Config {
	config := new_config()

	input := enter_config("a resource directory", default_resource_dir)
	config.set("resource_dir", input)
	input = enter_config("path to listening socket", default_soket)
	config.set("socket_path", input)

	write_config(path, config)
	return config
}

func write_config(path string, config Config) {
	var content = ""

	config_keys := config.keys()
	for i := 0; i < len(config_keys); i++ {
		content += config_keys[i] + " = " + config.Get(config_keys[i]) + "\n"
	}

	err := os.WriteFile(path, []byte(content), 0770)
	if err != nil {
		fmt.Println("Could not write to config file: " + path)
	}
}

func enter_config(description, default_value string) string {
	fmt.Printf("Please enter %s (%s)\n:", description, default_value)

	var input string
	fmt.Scanln(&input)
	if input == "" {
		input = default_value
	}
	return input
}
