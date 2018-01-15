package routing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cjsaylor/boxmeup-go/modules/database"
	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/middleware"
	"github.com/cjsaylor/boxmeup-go/modules/models"
	"github.com/cjsaylor/boxmeup-go/modules/users"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// IndexHandler serves the static page
func IndexHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Welcome!")
}

// HealthHandler serves up a health status.
func HealthHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNoContent)
}

// CreateLocationHandler will create a location from user input
// Expected body:
//   - name
//   - address
func CreateLocationHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	user, err := users.NewStore(db).ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Unable to find user to associate this location."})
		return
	}
	location := locations.Location{
		User:    user,
		Name:    req.PostFormValue("name"),
		Address: req.PostFormValue("address"),
	}
	err = locations.NewStore(db).Create(&location)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Unable to store location."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(map[string]int64{
		"id": location.ID,
	})
}

// UpdateLocationHandler will handle updating location based on user input
func UpdateLocationHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	locationModel := locations.NewStore(db)
	locationID, _ := strconv.Atoi(vars["id"])
	location, err := locationModel.ByID(int64(locationID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Location not found."})
		return
	}
	if userID != location.User.ID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to modify this location."})
		return
	}
	location.Name = req.PostFormValue("name")
	location.Address = req.PostFormValue("address")
	err = locationModel.Update(&location)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Failed to update location."})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

// DeleteLocationHandler will remove a location upon user request.
func DeleteLocationHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	locationModel := locations.NewStore(db)
	locationID, _ := strconv.Atoi(vars["id"])
	location, err := locationModel.ByID(int64(locationID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Location not found."})
		return
	}
	if userID != location.User.ID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to remove this location."})
		return
	}
	err = locationModel.Delete(int64(locationID))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Unable to remove location."})
	}
	res.WriteHeader(http.StatusNoContent)
}

// LocationsHandler will retrieve user locations
func LocationsHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	params := req.URL.Query()
	var limit models.QueryLimit
	page, _ := strconv.Atoi(params.Get("page"))
	limit.SetPage(page, locations.QueryLimit)
	locationModel := locations.NewStore(db)
	var sortField locations.SortableField
	if userSortField := params.Get("sort_field"); userSortField != "" {
		var err error
		sortField, err = locationModel.SortableFieldByName(userSortField)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Invalid sort field"})
			return
		}
	}
	user, err := users.NewStore(db).ByID(int64(userID))
	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "User not found."})
		return
	}
	sort := locationModel.GetSortBy(sortField, models.SortType(params.Get("sort_dir")))
	filter := locations.LocationFilter{
		User: user,
		IsAttachedToContainer: params.Get("is_attached_to_container") == "T",
	}
	response, err := locationModel.FilteredLocations(filter, sort, limit)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Unable to get locations."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(response)
}
