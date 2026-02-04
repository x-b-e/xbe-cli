package cli

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed client_routes.json
var clientRoutesFS embed.FS

var (
	clientRoutesOnce    sync.Once
	clientRoutesCatalog clientRouteCatalog
	clientRoutesErr     error
)

type clientRouteCatalog struct {
	Routes []clientRoute `json:"routes"`
}

type clientRoute struct {
	Key             string   `json:"key"`
	Path            string   `json:"path"`
	Params          []string `json:"params"`
	Action          bool     `json:"action"`
	TerminalSegment string   `json:"terminal_segment"`
	TerminalParam   string   `json:"terminal_param"`
}

func loadClientRoutes() (clientRouteCatalog, error) {
	clientRoutesOnce.Do(func() {
		data, err := clientRoutesFS.ReadFile("client_routes.json")
		if err != nil {
			clientRoutesErr = err
			return
		}
		if err := json.Unmarshal(data, &clientRoutesCatalog); err != nil {
			clientRoutesErr = err
			return
		}
		if len(clientRoutesCatalog.Routes) == 0 {
			clientRoutesErr = fmt.Errorf("client routes catalog is empty")
		}
	})
	return clientRoutesCatalog, clientRoutesErr
}
