package items

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/cjsaylor/boxmeup-go/config"
	"github.com/cjsaylor/boxmeup-go/database"
	"github.com/cjsaylor/boxmeup-go/middleware"
	"github.com/cjsaylor/boxmeup-go/models"
	"github.com/cjsaylor/boxmeup-go/modules/containers"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	chain "github.com/justinas/alice"
)

// Hook is the mechanism to plugin item module routes
type Hook struct{}

var routes = []config.Route{
	config.Route{
		Name:    "CreateContainerItem",
		Method:  "POST",
		Pattern: "/api/container/{id}/item",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(saveContainerItemHandler),
	},
	config.Route{
		Name:    "ModifyContainerItem",
		Method:  "PUT",
		Pattern: "/api/container/{id}/item/{item_id}",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(saveContainerItemHandler),
	},
	config.Route{
		Name:    "DeleteItems",
		Method:  "DELETE",
		Pattern: "/api/container/{id}/item/{item_id}",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(deleteContainerItemHandler),
	},
	config.Route{
		Name:    "DeleteItemsBulk",
		Method:  "POST",
		Pattern: "/api/container/item/bulk-delete",
		Handler: chain.New(middleware.AuthHandler).ThenFunc(deleteManyHandler),
	},
	config.Route{
		Name:    "Items",
		Method:  "GET",
		Pattern: "/api/container/{id}/item",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(containerItemsHandler),
	},
	config.Route{
		Name:    "Items",
		Method:  "GET",
		Pattern: "/api/item/search",
		Handler: chain.New(middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(searchItemHandler),
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

type bulkDeleteID struct {
	IDs []int64 `json:"ids"`
}

type bulkItemRetrieval struct {
	item ContainerItem
	err  error
}

type bulkItemRetrievals []bulkItemRetrieval

func (b *bulkItemRetrievals) anyErrors() bool {
	for _, resp := range *b {
		if resp.err != nil {
			return true
		}
	}
	return false
}

func (b *bulkItemRetrievals) items() *ContainerItems {
	items := ContainerItems{}
	for _, resp := range *b {
		items = append(items, resp.item)
	}
	return &items
}

// ContainerItemsHandler is an interface into items of a container
// @todo Consider syncing some of the non-related queries to go routines
func containerItemsHandler(res http.ResponseWriter, req *http.Request) {
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
	limit.SetPage(page, QueryLimit)
	itemModel := NewStore(db)
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

func searchItemHandler(res http.ResponseWriter, req *http.Request) {
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
	itemModel := NewStore(db)
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

// SaveContainerItemHandler allows creation of a container from a POST method
// Expected body:
//   body
//   quantity
func saveContainerItemHandler(res http.ResponseWriter, req *http.Request) {
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
	itemModel := NewStore(db)
	quantity, _ := strconv.Atoi(req.PostFormValue("quantity"))
	var item ContainerItem
	if _, ok := vars["item_id"]; ok {
		itemID, _ := strconv.Atoi(vars["item_id"])
		item, err = itemModel.ByID(int64(itemID))
		if err != nil {
			jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: "Unable to retrieve item to modify."})
		}
	} else {
		item = ContainerItem{
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
func deleteContainerItemHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	jsonOut := json.NewEncoder(res)
	vars := mux.Vars(req)
	itemID, _ := strconv.Atoi(vars["item_id"])
	itemModel := NewStore(db)
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

func deleteManyHandler(res http.ResponseWriter, req *http.Request) {
	db, _ := database.GetDBResource()
	defer db.Close()
	jsonOut := json.NewEncoder(res)
	userID := int64(req.Context().Value(middleware.UserContextKey).(jwt.MapClaims)["id"].(float64))
	decoder := json.NewDecoder(req.Body)
	itemStore := NewStore(db)
	var bulkOptions bulkDeleteID
	decoder.Decode(&bulkOptions)
	var wg sync.WaitGroup
	retrieve := make(chan bulkItemRetrieval, len(bulkOptions.IDs))
	for _, id := range bulkOptions.IDs {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			resp := bulkItemRetrieval{}
			resp.item, resp.err = itemStore.ByID(id)
			retrieve <- resp
		}(id)
	}
	wg.Wait()
	close(retrieve)
	var itemsRetrieved bulkItemRetrievals
	for resp := range retrieve {
		itemsRetrieved = append(itemsRetrieved, resp)
	}
	if itemsRetrieved.anyErrors() {
		res.WriteHeader(http.StatusNotFound)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -1, Text: "Some or all of the items could not be retrieved"})
		return
	}
	users := itemsRetrieved.items().ExtractUsers()
	if len(users) > 1 || users[0].ID != userID {
		res.WriteHeader(http.StatusForbidden)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -2, Text: "Not authorized to delete some or all of the items"})
		return
	}
	err := itemStore.DeleteMany(*itemsRetrieved.items())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		jsonOut.Encode(middleware.JsonErrorResponse{Code: -3, Text: err.Error()})
		return
	}
	res.WriteHeader(http.StatusNoContent)
}
