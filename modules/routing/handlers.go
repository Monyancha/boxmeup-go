package routing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cjsaylor/boxmeup-go/modules/config"
	"github.com/cjsaylor/boxmeup-go/modules/containers"
	"github.com/cjsaylor/boxmeup-go/modules/database"
	"github.com/cjsaylor/boxmeup-go/modules/items"
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

// LoginHandler authenticates via email and password
func LoginHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()

	token, err := users.NewStore(db).Login(
		users.AuthConfig{
			LegacySalt: config.Config.LegacySalt,
			JWTSecret:  config.Config.JWTSecret,
		},
		req.PostFormValue("email"),
		req.PostFormValue("password"))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Authentication failure."})
	} else {
		expiration := time.Now().Add(14 * 24 * time.Hour)
		cookie := http.Cookie{
			Name:     middleware.SessionName,
			Value:    token,
			Expires:  expiration,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(res, &cookie)
		res.WriteHeader(http.StatusOK)
		jsonOut.Encode(map[string]string{
			"token": token,
		})
	}
}

func LogoutHandler(res http.ResponseWriter, req *http.Request) {
	cookie := http.Cookie{
		Name:     middleware.SessionName,
		Value:    "",
		Expires:  time.Now().Add(-100 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(res, &cookie)
	res.WriteHeader(http.StatusNoContent)
}

// RegisterHandler creates new users.
func RegisterHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()

	email := req.PostFormValue("email")
	password := req.PostFormValue("password")
	id, err := users.NewStore(db).Register(
		users.AuthConfig{
			LegacySalt: config.Config.LegacySalt,
			JWTSecret:  config.Config.JWTSecret,
		},
		email,
		password)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: err.Error()})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(map[string]int64{
		"id": id,
	})
}

// UserHandler returns basic user information of current user.
func UserHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	user, err := users.NewStore(db).ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "User specified not found."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(user)
}

// CreateContainerHandler allows creation of a container from a POST method
// Expected body:
//   name
//   location_id (optional)
func CreateContainerHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	user, err := users.NewStore(db).ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "User specified not found."})
		return
	}
	record := containers.NewRecord(&user)
	record.Name = req.PostFormValue("name")
	if userLocationID := req.PostFormValue("location_id"); userLocationID != "" {
		locationID, _ := strconv.Atoi(userLocationID)
		location, err := locations.NewStore(db).ByID(int64(locationID))
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			jsonOut.Encode(middleware.JsonErrorResponse{Code: -5, Text: "Location not found."})
			return
		} else if location.User.ID != userID {
			res.WriteHeader(http.StatusForbidden)
			jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Not allowed to attach supplied location to this container."})
			return
		}
		record.SetLocation(&location)
	} else {
		record.SetLocation(nil)
	}
	err = containers.NewStore(db).Create(&record)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Failed to create the container."})
	} else {
		res.WriteHeader(http.StatusOK)
		jsonOut.Encode(map[string]int64{
			"id": record.ID,
		})
	}
}

// UpdateContainerHandler exposes updating a container
// @todo consider a new endpoint for just location attachment/detachment and remove location editing here
// -> PUT /api/container/<id>/location/<location_id>
// -> DELETE /api/container/<id>/location
func UpdateContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	containerModel := containers.NewStore(db)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containerModel.ByID(int64(containerID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to edit this container."})
		return
	}
	record := container.ToRecord()
	record.Name = req.PostFormValue("name")
	if userLocationID := req.PostFormValue("location_id"); userLocationID != "" {
		locationID, _ := strconv.Atoi(userLocationID)
		location, err := locations.NewStore(db).ByID(int64(locationID))
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			jsonOut.Encode(middleware.JsonErrorResponse{Code: -5, Text: "Location not found."})
			return
		} else if location.User.ID != userID {
			res.WriteHeader(http.StatusForbidden)
			jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Not allowed to attach supplied location to this container."})
			return
		}
		record.SetLocation(&location)
	} else {
		record.SetLocation(nil)
	}
	err = containerModel.Update(&record)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -4, Text: err.Error()})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

// DeleteContainerHandler removes a container on request of the user
func DeleteContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	containerModel := containers.NewStore(db)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containerModel.ByID(int64(containerID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to edit this container."})
		return
	}
	err = containerModel.Delete(int64(containerID))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Error deleting container."})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

// ContainerHandler gets a specific container by ID
func ContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containers.NewStore(db).ByID(int64(containerID))
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to view this container."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(container)
}

// ContainersHandler gets all user containers
func ContainersHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	userModel := users.NewStore(db)
	user, err := userModel.ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Unable to get user information."})
		return
	}
	params := req.URL.Query()
	var limit models.QueryLimit
	page, _ := strconv.Atoi(params.Get("page"))
	limit.SetPage(page, containers.QueryLimit)
	containerModel := containers.NewStore(db)
	sort := containerModel.GetSortBy(params.Get("sort_field"), models.SortType(params.Get("sort_dir")))
	filter := containers.ContainerFilter{
		User:        user,
		LocationIDs: params["location_id"],
	}
	response, err := containerModel.FilteredContainers(filter, sort, limit)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Unable to retrieve containers."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(response)
}

// SaveContainerItemHandler allows creation of a container from a POST method
// Expected body:
//   body
//   quantity
func SaveContainerItemHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	vars := mux.Vars(req)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containers.NewStore(db).ByID(int64(containerID))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Failed to retrieve the container."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to modify this container."})
		return
	}
	itemModel := items.NewStore(db)
	quantity, _ := strconv.Atoi(req.PostFormValue("quantity"))
	var item items.ContainerItem
	if _, ok := vars["item_id"]; ok {
		itemID, _ := strconv.Atoi(vars["item_id"])
		item, err = itemModel.ByID(int64(itemID))
		if err != nil {
			jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Unable to retrieve item to modify."})
		}
	} else {
		item = items.ContainerItem{
			Container: &container,
		}
	}
	if quantity > 0 {
		item.Quantity = quantity
	}
	if body := req.PostFormValue("body"); body != "" {
		item.Body = body
	}
	if _, ok := vars["item_id"]; ok {
		itemID, _ := strconv.Atoi(vars["item_id"])
		item.ID = int64(itemID)
		err = itemModel.Update(item)
	} else {
		err = itemModel.Create(&item)
	}
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -4, Text: "Unable to create container item"})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(map[string]int64{
		"id": item.ID,
	})
}

// DeleteContainerItemHandler will remove an item from a container and update the container count.
func DeleteContainerItemHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	vars := mux.Vars(req)
	itemID, _ := strconv.Atoi(vars["item_id"])
	itemModel := items.NewStore(db)
	item, err := itemModel.ByID(int64(itemID))
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Item not found."})
		return
	}
	if item.Container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to delete this item."})
		return
	}
	err = itemModel.Delete(item)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Unable to delete this item."})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}

// ContainerItemsHandler is an interface into items of a container
// @todo Consider syncing some of the non-related queries to go routines
func ContainerItemsHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	vars := mux.Vars(req)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := containers.NewStore(db).ByID(int64(containerID))
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Container not found."})
		return
	}
	if container.User.ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not allowed to view items in this container."})
		return
	}
	params := req.URL.Query()
	var limit models.QueryLimit
	page, _ := strconv.Atoi(params.Get("page"))
	limit.SetPage(page, items.QueryLimit)
	itemModel := items.NewStore(db)
	sort := itemModel.GetSortBy(params.Get("sort_field"), models.SortType(params.Get("sort_dir")))
	response, err := itemModel.GetContainerItems(&container, sort, limit)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Unable to retrieve container items."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(response)
}

func SearchItemHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	params := req.URL.Query()
	var limit models.QueryLimit
	term := params.Get("term")
	jsonOut := json.NewEncoder(res)
	if term == "" {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Must provide a search term."})
		return
	}
	page, _ := strconv.Atoi(params.Get("page"))
	limit.SetPage(page, containers.QueryLimit)
	itemModel := items.NewStore(db)
	sort := itemModel.GetSortBy(params.Get("sort_field"), models.SortType(params.Get("sort_dir")))
	response, err := itemModel.SearchItems(int64(userID), term, sort, limit)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Unable to retrieve items."})
		return
	}
	res.WriteHeader(http.StatusOK)
	jsonOut.Encode(response)
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
