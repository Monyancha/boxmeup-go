package items

import (
	"time"

	"github.com/cjsaylor/boxmeup-go/modules/containers"
	"github.com/cjsaylor/boxmeup-go/modules/users"
)

// ContainerItem represents a single item in a container
type ContainerItem struct {
	ID        int64                 `json:"id"`
	Container *containers.Container `json:"-"`
	UUID      string                `json:"uuid"`
	Body      string                `json:"body"`
	Quantity  int                   `json:"quantity"`
	Created   time.Time             `json:"created"`
	Modified  time.Time             `json:"modified"`
}

// ContainerItems is a collection of container items.
type ContainerItems []ContainerItem

// ExtractUsers from arbitrary container items
func (i *ContainerItems) ExtractUsers() []users.User {
	ids := make(map[int64]bool)
	users := make([]users.User, 0)
	for _, item := range *i {
		if !ids[item.Container.User.ID] {
			ids[item.Container.User.ID] = true
			users = append(users, item.Container.User)
		}
	}
	return users
}

// ExtractContainers from arbitrary container items
func (i *ContainerItems) ExtractContainers() []containers.Container {
	ids := make(map[int64]bool)
	containers := make([]containers.Container, 0)
	for _, item := range *i {
		if !ids[item.ID] {
			ids[item.ID] = true
			containers = append(containers, *item.Container)
		}
	}
	return containers
}
