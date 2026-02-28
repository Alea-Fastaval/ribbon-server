package translations

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dreamspawn/ribbon-server/config"
)

var tranlations_dir string
var fallback_lang string
var languages []string

func ConfigReady() {
	tranlations_dir = config.Get("resource_dir") + "translations/"
	fallback_lang = config.Get("fallback_lang")
}

var translations map[string]map[string]map[string]string

func LoadAll() {
	for _, lang := range GetLanguages() {
		files, err := os.ReadDir(tranlations_dir + lang + "/")
		if err != nil {
			fmt.Printf("Could not read content of folder %s\n", tranlations_dir+lang+"/")
			panic(err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			load(file.Name(), lang)
		}
	}
}

func load(file_name string, lang string) error {
	// Initialize Map
	if translations == nil {
		translations = make(map[string]map[string]map[string]string)
	}

	if translations[lang] == nil {
		translations[lang] = make(map[string]map[string]string)
	}

	// open translation file
	path := tranlations_dir + lang + "/" + file_name
	file, err := os.Open(path)
	if err != nil {
		// Try fallback language file
		path = tranlations_dir + fallback_lang + "/" + file_name
		file, err = os.Open(path)
		if err != nil {
			return err
		}
	}
	translation_set := make(map[string]string)

	reader := bufio.NewReader(file)
	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		line := string(bytes)

		// Allow comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line, err = strconv.Unquote("\"" + line + "\"")
		key, translation, _ := strings.Cut(line, " ")
		translation_set[key] = translation
	}

	parts := strings.Split(file_name, ".")
	translations[lang][parts[0]] = translation_set

	return nil
}

func Get(lang string, set string, key string) string {
	translation, found := translations[lang][set][key]
	if !found {
		translation = "Missing translation (" + key + ")"
		log.Output(2, "Missing translation for Lang:"+lang+" Set:"+set+" Key:"+key+"\n")
	}

	return translation
}

func GetSet(lang string, set string) map[string]string {
	return translations[lang][set]
}

func GetLanguages() []string {
	if len(languages) == 0 {
		loadLanguages()
	}

	return languages
}

func loadLanguages() {
	files, err := os.ReadDir(tranlations_dir)
	if err != nil {
		fmt.Printf("Could not read content of folder %s\n", tranlations_dir)
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		languages = append(languages, file.Name())
	}
}
