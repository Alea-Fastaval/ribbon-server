package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dreamspawn/ribbon-server/database"
	"github.com/dreamspawn/ribbon-server/server/session"
)

// /api/orders
func ordersAPI(sub_path string, vars url.Values, request http.Request) (any, error) {
	current_session := session.Check(request)
	if current_session == nil {
		api_error("Not logged in", nil)
	}

	user := current_session.GetUser()

	//--------------------------------------------------------------------------------------------------------------------
	// GET
	//--------------------------------------------------------------------------------------------------------------------
	if request.Method == "GET" {
		var orders any
		var err error
		if user.IsAdmin {
			orders, err = database.GetAllOrders()
		} else {
			orders, err = database.GetOrders(user.ID)
		}

		if err != nil {
			api_error("could not load orders from database", nil)
		}

		return orders, nil
	}

	//--------------------------------------------------------------------------------------------------------------------
	// POST
	//--------------------------------------------------------------------------------------------------------------------
	if request.Method == "POST" {

		var ribbon_id uint64
		if ribbon, ok := vars["ribbon"]; ok {
			ribbon_id, _ = strconv.ParseUint(ribbon[0], 10, 32)
		} else {
			api_error("missing parameter: ribbon", nil)
		}

		values := make(map[string]uint)
		keys := []string{
			"grunt",
			"leader",
			"second",
			"position",
		}

		for _, key := range keys {
			if raw, ok := vars[key]; ok {
				parsed, err := strconv.ParseUint(raw[0], 10, 32)
				if err == nil {
					values[key] = uint(parsed)
				}
			}
		}

		if len(values) == 0 {
			api_error("missing data for ribbon order", nil)
		}

		err := database.SetOrder(
			uint(user.ID),
			uint(ribbon_id),
			values,
		)

		if err != nil {
			api_error(fmt.Sprintf("Failed to create new ribbon order with values %+v\n", vars), err)
		}

		return map[string]string{
			"status":  "success",
			"message": "ribbon order set",
		}, nil
	}

	return nil, fmt.Errorf("endpoint not implemented for method %s", request.Method)
}
