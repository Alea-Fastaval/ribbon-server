package database

type Category struct {
	id         uint
	foreground string
	background string
	ordering   uint
}

func GetCategories() ([]Category, error) {
	statement := "SELECT * FROM categories"
	rows, err := db.Query(statement)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	var result []Category
	for rows.Next() {
		category := Category{}

		err := rows.Scan(&category.id, &category.foreground, &category.background, &category.ordering)
		if err != nil {
			return nil, db_error(statement, nil, err)
		}

		result = append(result, category)
	}

	return result, nil
}
