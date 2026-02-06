package cli

import (
	"embed"
	"encoding/json"
	"sync"
)

//go:embed client_route_docs.json
var clientRouteDocsFS embed.FS

var (
	clientRouteDocsOnce    sync.Once
	clientRouteDocsCatalog clientRouteDocs
	clientRouteDocsErr     error
)

type clientRouteDocs struct {
	Routes map[string]clientRouteDoc `json:"routes"`
}

type clientRouteDoc struct {
	Summary             string   `json:"summary,omitempty"`
	QueryParams         []string `json:"query_params,omitempty"`
	RequiredQueryParams []string `json:"required_query_params,omitempty"`
	ResourceValues      []string `json:"resource_values,omitempty"`
	Examples            []string `json:"examples,omitempty"`
	Notes               []string `json:"notes,omitempty"`
	SourcePaths         []string `json:"source_paths,omitempty"`
}

func loadClientRouteDocs() (clientRouteDocs, error) {
	clientRouteDocsOnce.Do(func() {
		data, err := clientRouteDocsFS.ReadFile("client_route_docs.json")
		if err != nil {
			clientRouteDocsErr = err
			return
		}
		if err := json.Unmarshal(data, &clientRouteDocsCatalog); err != nil {
			clientRouteDocsErr = err
			return
		}
		if clientRouteDocsCatalog.Routes == nil {
			clientRouteDocsCatalog.Routes = map[string]clientRouteDoc{}
		}
	})
	return clientRouteDocsCatalog, clientRouteDocsErr
}
