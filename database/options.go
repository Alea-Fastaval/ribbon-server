package database

func GetOptions() (map[string]any, error) {
	query := "SELECT * FROM options"
	result, err := Query(query, []any{})

	options := make(map[string]any)
	for _, row := range result {
		options[row["name"].(string)] = row["value"]
	}

	return options, err
}

func SetOption(name, value string) error {
	query := "UPDATE options SET value = ? WHERE name = ?"
	_, err := Exec(query, []any{value, name})
	return err
}
