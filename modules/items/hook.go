package items

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/cjsaylor/boxmeup-go/modules/database"
	"github.com/cjsaylor/boxmeup-go/modules/middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

// Hook is the mechanism to plugin item module routes
type Hook struct{}

// Apply hooks related to items
func (h Hook) Apply(router *mux.Router) {
	router.
		Methods("POST").
		Path("/api/container/item/bulk-delete").
		Name("DeleteItemsBulk").
		Handler(alice.New(middleware.LogHandler, middleware.AuthHandler).ThenFunc(deleteManyHandler))
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
