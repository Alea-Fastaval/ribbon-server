package database

type Ribbon struct {
	ID       uint
	Category uint
	Glyph    uint
	NoWings  bool
	Ordering uint
}

// ----------------------------------------------------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------------------------------------------------
func CreateRibbon(category uint, glyph uint, no_wings bool) (*Ribbon, error) {
	var ribbon_count uint
	statement := "SELECT COUNT(*) FROM ribbons WHERE category_id = ?"
	row := db.QueryRow(statement, category)
	err := row.Scan(&ribbon_count)
	if err != nil {
		args := []any{
			category,
		}
		return nil, db_error(statement, args, err)
	}

	statement = "INSERT INTO ribbons(category_id, glyph_id, no_wings, ordering) VALUES(?,?,?,?)"
	result, err := db.Exec(statement, category, glyph, no_wings, ribbon_count)
	if err != nil {
		args := []any{
			category,
			glyph,
			no_wings,
			ribbon_count,
		}
		return nil, db_error(statement, args, err)
	}

	id, err := result.LastInsertId()
	new_ribbon := Ribbon{
		uint(id),
		category,
		glyph,
		no_wings,
		ribbon_count,
	}

	return &new_ribbon, err
}

// ----------------------------------------------------------------------------------------------------------------------
// Read
// ----------------------------------------------------------------------------------------------------------------------
func GetRibbons() ([]Ribbon, error) {
	statement := "SELECT * FROM ribbons ORDER BY ordering"
	rows, err := db.Query(statement)
	if err != nil {
		return nil, err
	}

	var result []Ribbon

	for rows.Next() {
		ribbon := Ribbon{}

		err := rows.Scan(
			&ribbon.ID,
			&ribbon.Category,
			&ribbon.Glyph,
			&ribbon.NoWings,
			&ribbon.Ordering,
		)

		if err != nil {
			return nil, db_error(statement, nil, err)
		}

		result = append(result, ribbon)
	}

	return result, nil
}

// ----------------------------------------------------------------------------------------------------------------------
// Delete
// ----------------------------------------------------------------------------------------------------------------------
func DeleteRibbon(id uint) error {
	query := "DELETE FROM ribbons WHERE id = ?"
	_, err := Exec(query, []any{id})
	return err
}
