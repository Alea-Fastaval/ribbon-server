package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var default_resource_dir = "/var/www/ribbon/"
var default_soket = "/var/run/ribbon.sock"

var values map[string]string

func keys() []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	return keys
}

func set(key, value string) {
	values[key] = value
}

func Get(key string) string {
	return values[key]
}

func LoadConfig(path string) error {
	values = make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(file)

	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		tokens := strings.Split(string(bytes), "=")
		set(strings.Trim(tokens[0], " "), strings.Trim(tokens[1], " "))
	}

	return nil
}

func CreateConfig(path string) {
	values = make(map[string]string)

	input := enter_config("a resource directory", default_resource_dir)
	set("resource_dir", input)
	input = enter_config("path to listening socket", default_soket)
	set("socket_path", input)

	write_config(path)
}

func write_config(path string) {
	fmt.Printf("Saving config to %s", path)
	var content = ""

	config_keys := keys()
	for i := 0; i < len(config_keys); i++ {
		content += config_keys[i] + " = " + Get(config_keys[i]) + "\n"
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
