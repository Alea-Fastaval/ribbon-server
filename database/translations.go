package database

import "strings"

func AddTranslation(lang string, label string, value string) error {
	statement := "INSERT INTO translations (lang, label, string) VALUES(?,?,?)"
	_, err := db.Exec(statement, lang, label, value)

	return err
}

func DeleteTranslation(lang, label string) error {
	statement := "DELETE FROM translations WHERE lang = ? AND label = ?"
	_, err := db.Exec(statement, lang, label)

	return err
}

func GetTranslation(lang string, key string) (map[string]interface{}, error) {
	key = strings.ReplaceAll(key, "*", "%")
	statement := "SELECT label, string FROM translations WHERE lang = ? AND label LIKE ?"
	rows, err := db.Query(statement, lang, key)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	result := make(map[string]interface{})
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, db_error(statement, nil, err)
		}

		recursive_insert(result, key, value)
	}

	return result, nil
}

func recursive_insert(collection map[string]interface{}, key string, value string) {
	before, after, found := strings.Cut(key, ".")
	if !found {
		collection[before] = value
	}

	_, ok := collection[before]
	if !ok {
		collection[before] = make(map[string]interface{})
	}

	sub_collection, ok := collection[before].(map[string]interface{})
	if ok {
		recursive_insert(sub_collection, after, value)
	}
}
