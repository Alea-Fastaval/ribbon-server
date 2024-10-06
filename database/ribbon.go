package database

type Ribbon struct {
	ID       uint
	Category uint
	Glyph    uint
	NoWings  bool
}

func CreateRibbon(category uint, glyph uint, no_wings bool) (*Ribbon, error) {
	var ribbon_count int
	statement := "SELECT COUNT(*) FROM ribbons WHERE category_id = ?"
	row := db.QueryRow(statement, category)
	err := row.Scan(&ribbon_count)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	statement = "INSERT INTO ribbons(category_id, glyph_id, no_wings, ordering) VALUES(?,?,?,?)"
	result, err := db.Exec(statement, category, glyph, no_wings, ribbon_count)
	if err != nil {
		return nil, db_error(statement, nil, err)
	}

	id, err := result.LastInsertId()
	new_ribbon := Ribbon{
		uint(id),
		category,
		glyph,
		no_wings,
	}

	return &new_ribbon, err
}
