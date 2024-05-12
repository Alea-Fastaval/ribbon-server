package api

import (
	"net/url"
	"strings"

	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/translations"
)

func translationsAPI(sub_path string, vars url.Values, method string) (any, error) {
	if sub_path == "languages" {
		return translations.GetLanguages(), nil
	}

	lang, key, _ := strings.Cut(sub_path, "/")
	if file, found := strings.CutSuffix(key, "/"); found {
		return translations.GetSet(lang, file), nil
	}

	return database.GetTranslation(lang, key)
}
