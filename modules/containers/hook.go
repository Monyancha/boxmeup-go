package containers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cjsaylor/boxmeup-go/config"
	"github.com/cjsaylor/boxmeup-go/database"
	"github.com/cjsaylor/boxmeup-go/middleware"
	"github.com/cjsaylor/boxmeup-go/models"
	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/users"
	"github.com/gorilla/mux"
	chain "github.com/justinas/alice"
)

// Hook is the mechanism to plugin item module routes
type Hook struct{}

var routes = []config.Route{
	config.Route{
		Name:    "CreateContainer",
		Method:  "POST",
		Pattern: "/api/container",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(createContainerHandler),
	},
	config.Route{
		Name:    "UpdateContainer",
		Method:  "PUT",
		Pattern: "/api/container/{id}",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(updateContainerHandler),
	},
	config.Route{
		Name:    "DeleteContainer",
		Method:  "DELETE",
		Pattern: "/api/container/{id}",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(deleteContainerHandler),
	},
	config.Route{
		Name:    "Container",
		Method:  "GET",
		Pattern: "/api/container/{id}",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(containerHandler),
	},
	config.Route{
		Name:    "Containers",
		Method:  "GET",
		Pattern: "/api/container",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(containersHandler),
	},
}

// Apply hooks related to items
func (h Hook) Apply(router *mux.Router) {
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.Handler)
	}
}

// createContainerHandler allows creation of a container from a POST method
// Expected body:
//   name
//   location_id (optional)
func createContainerHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := middleware.UserIDFromRequest(req)
	user, err := users.NewStore(db).ByID(userID)
	jsonOut := json.NewEncoder(res)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "User specified not found."})
		return
	}
	record := NewRecord(&user)
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
	err = NewStore(db).Create(&record)
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

// updateContainerHandler exposes updating a container
// @todo consider a new endpoint for just location attachment/detachment and remove location editing here
// -> PUT /api/container/<id>/location/<location_id>
// -> DELETE /api/container/<id>/location
func updateContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := middleware.UserIDFromRequest(req)
	containerModel := NewStore(db)
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

// deleteContainerHandler removes a container on request of the user
func deleteContainerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := middleware.UserIDFromRequest(req)
	containerModel := NewStore(db)
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

// containerHandler gets a specific container by ID
func containerHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := middleware.UserIDFromRequest(req)
	containerID, _ := strconv.Atoi(vars["id"])
	container, err := NewStore(db).ByID(int64(containerID))
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

// containersHandler gets all user containers
func containersHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := middleware.UserIDFromRequest(req)
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
	limit.SetPage(page, QueryLimit)
	containerModel := NewStore(db)
	sort := containerModel.GetSortBy(params.Get("sort_field"), models.SortType(params.Get("sort_dir")))
	filter := ContainerFilter{
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
