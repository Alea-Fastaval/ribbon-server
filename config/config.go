package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var defaults = map[string][2]string{
	"resource_dir": {
		"a resource directory",
		"/var/www/ribbon/",
	},
	"socket_path": {
		"path to listening socket",
		"/var/run/ribbon.sock",
	},
	"admin_slug": {
		"prefix for admin pages",
		"backroom",
	},
	"fallback_lang": {
		"default language code",
		"da",
	},
}

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
		set(strings.Trim(tokens[0], " \t"), strings.Trim(tokens[1], " \t"))
	}

	return nil
}

func CreateConfig(path string) {
	values = make(map[string]string)

	for key, info := range defaults {
		input := enter_config(info[0], info[1])
		set(key, input)
	}

	write_config(path)
}

func write_config(path string) {
	fmt.Printf("Saving config to %s\n", path)
	var content = ""

	for _, key := range keys() {
		content += key + " = " + Get(key) + "\n"
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
		return default_value
	}
	return input
}
