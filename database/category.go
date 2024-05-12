package database

type Category struct {
	ID         uint
	Background string
	Stripes    string
	Glyph      string
	Ordering   uint
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

		err := rows.Scan(&category.ID, &category.Background, &category.Stripes, &category.Glyph, &category.Ordering)
		if err != nil {
			return nil, db_error(statement, nil, err)
		}

		result = append(result, category)
	}

	return result, nil
}

func CreateCategory(background, stripes, glyph string) (*Category, error) {
	var category_count int
	statement := "SELECT COUNT(*) FROM categories"
	row := db.QueryRow(statement)
	err := row.Scan(&category_count)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	statement = "INSERT INTO categories(background, stripes, glyph, ordering) VALUES(?,?,?,?)"
	result, err := db.Exec(statement, background, stripes, glyph, category_count)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	id, err := result.LastInsertId()

	new_category := Category{
		uint(id),
		background,
		stripes,
		glyph,
		uint(category_count),
	}
	return &new_category, err
}
