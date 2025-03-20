package database

import (
	"strings"
)

func SetOrder(uid, ribbon uint, values map[string]uint) error {
	query := "SELECT * FROM ribbon_orders WHERE user_id = ? AND ribbon_id = ?"
	order, _ := Query(query, []any{uid, ribbon})

	if len(order) != 0 {
		return update(uid, ribbon, values)
	}

	return insert(uid, ribbon, values)
}

func insert(uid, ribbon uint, values map[string]uint) error {
	// Set default fro missing values
	args := []any{uid, ribbon}
	keys := []string{
		"grunt",
		"leader",
		"second",
	}
	for _, key := range keys {
		if value, ok := values[key]; ok {
			args = append(args, value)
		} else {
			args = append(args, 0)
		}
	}

	// Find position for new ribbon
	query := "SELECT COUNT(*) as count FROM ribbon_orders WHERE user_id = ?"
	result_count, err := Query(query, []any{uid})
	if err != nil {
		return err
	}
	poistion := result_count[0]["count"].(int64) + 1
	args = append(args, poistion)

	// Insert ribbon order
	query = "INSERT INTO ribbon_orders (user_id, ribbon_id, grunt, second, leader, position) VALUES(?,?,?,?,?,?)"
	_, err = Exec(query, args)

	return err
}

func update(uid, ribbon uint, values map[string]uint) error {
	query := "UPDATE ribbon_orders SET "

	var args []any
	var value_queries []string
	for key, value := range values {
		value_queries = append(value_queries, key+" = ?")
		args = append(args, value)
	}

	query += strings.Join(value_queries, ", ")
	query += " WHERE user_id = ? AND ribbon_id = ?"
	args = append(args, uid, ribbon)

	_, err := Exec(query, args)

	return err
}

func GetOrders(uid uint) (map[uint]map[string]uint, error) {
	return make(map[uint]map[string]uint), nil
}

func GetAllOrders() (map[uint]map[string]uint, error) {
	return make(map[uint]map[string]uint), nil
}
