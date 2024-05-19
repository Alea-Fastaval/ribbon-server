package database

type Category struct {
	ID         uint
	Background string
	Stripes    string
	Glyph      string
	Wing1      string
	Wing2      string
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

		err := rows.Scan(
			&category.ID,
			&category.Background,
			&category.Stripes,
			&category.Glyph,
			&category.Wing1,
			&category.Wing2,
			&category.Ordering,
		)
		if err != nil {
			return nil, db_error(statement, nil, err)
		}

		result = append(result, category)
	}

	return result, nil
}

func CreateCategory(background, stripes, glyph, wing1, wing2 string) (*Category, error) {
	var category_count int
	statement := "SELECT COUNT(*) FROM categories"
	row := db.QueryRow(statement)
	err := row.Scan(&category_count)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	statement = "INSERT INTO categories(background, stripes, glyph, wing1, wing2, ordering) VALUES(?,?,?,?,?,?)"
	result, err := db.Exec(statement, background, stripes, glyph, wing1, wing2, category_count)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	id, err := result.LastInsertId()

	new_category := Category{
		uint(id),
		background,
		stripes,
		glyph,
		wing1,
		wing2,
		uint(category_count),
	}
	return &new_category, err
}
