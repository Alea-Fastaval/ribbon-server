package translations

import (
	"bufio"
	"os"
	"strings"

	"github.com/dreamspawn/ribbon-server/config"
)

var tranlations_dir string
var fallback_lang string

func ConfigReady() {
	tranlations_dir = config.Get("resource_dir") + "translations/"
	fallback_lang = config.Get("fallback_lang")
}

var translations map[string]map[string]map[string]string

func Load(name string, lang string) error {
	// Initialize Map
	if translations == nil {
		translations = make(map[string]map[string]map[string]string)
	}

	if translations[lang] == nil {
		translations[lang] = make(map[string]map[string]string)
	}

	// open translation file
	path := tranlations_dir + lang + "/" + name + ".txt"
	file, err := os.Open(path)
	if err != nil {
		// Try fallback language file
		path = tranlations_dir + fallback_lang + "/" + name + ".txt"
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

		key, translation, _ := strings.Cut(string(bytes), " ")
		translation_set[key] = translation
	}

	translations[lang][name] = translation_set

	return nil
}

func Get(lang string, set string, key string) string {
	return translations[lang][set][key]
}

func GetSet(lang string, set string) map[string]string {
	return translations[lang][set]
}
