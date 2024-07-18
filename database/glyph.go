package database

type Glyph struct {
	ID   uint
	File string
}

func CreateGlyph(file_name string) (*Glyph, error) {
	statement := "INSERT INTO glyphs(file_name) VALUES(?)"
	result, err := db.Exec(statement, file_name)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	id, err := result.LastInsertId()

	new_glyph := Glyph{
		uint(id),
		file_name,
	}

	return &new_glyph, err
}

func GetGlyphs() (map[uint]Glyph, error) {
	statement := "SELECT * FROM glyphs"
	rows, err := db.Query(statement)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	result := make(map[uint]Glyph)
	for rows.Next() {
		glyph := Glyph{}

		err := rows.Scan(
			&glyph.ID,
			&glyph.File,
		)
		if err != nil {
			return nil, db_error(statement, nil, err)
		}

		result[glyph.ID] = glyph
	}

	return result, nil
}

func GetGlyph(id uint) (*Glyph, error) {
	statement := "SELECT * FROM glyphs WHERE id = ?"
	row := db.QueryRow(statement, id)

	glyph := Glyph{}

	err := row.Scan(
		&glyph.ID,
		&glyph.File,
	)

	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	return &glyph, nil
}
