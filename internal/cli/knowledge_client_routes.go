package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeClientRouteRow struct {
	Key             string   `json:"key"`
	Path            string   `json:"path"`
	Params          []string `json:"params,omitempty"`
	Action          bool     `json:"action"`
	TerminalSegment string   `json:"terminal_segment,omitempty"`
	TerminalParam   string   `json:"terminal_param,omitempty"`
	Resources       []string `json:"resources,omitempty"`
}

func newKnowledgeClientRoutesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client-routes",
		Short: "List client app routes",
		Long: `List client app routes from the embedded client router catalog.

Use this to discover view URLs, debug --client-url output, and understand
which client views reference a given resource or parameter set.`,
		RunE: runKnowledgeClientRoutes,
		Example: `  # List client routes (default limit applies)
  xbe knowledge client-routes

  # List all routes (raise the limit)
  xbe knowledge client-routes --limit 2000

  # Search by keyword
  xbe knowledge client-routes --query job-production-plans

  # Routes that reference a resource
  xbe knowledge client-routes --resource projects

  # Only action routes
  xbe knowledge client-routes --action true

  # Show full client URLs
  xbe knowledge client-routes --resource users --full-url`,
	}
	cmd.Flags().String("query", "", "Substring filter for key, path, params, terminal, or resource")
	cmd.Flags().String("resource", "", "Only routes that reference a resource (comma-separated ok)")
	cmd.Flags().String("path", "", "Substring filter for route path")
	cmd.Flags().String("key", "", "Substring filter for route key")
	cmd.Flags().String("param", "", "Substring filter for param name")
	cmd.Flags().String("terminal-segment", "", "Substring filter for terminal segment")
	cmd.Flags().String("terminal-param", "", "Substring filter for terminal param")
	cmd.Flags().String("action", "", "Filter by action routes (true/false)")
	cmd.Flags().Bool("full-url", false, "Render full client URLs instead of route paths")
	return cmd
}

func runKnowledgeClientRoutes(cmd *cobra.Command, _ []string) error {
	catalog, err := loadClientRoutes()
	if err != nil {
		return err
	}

	resourceMap, err := loadResourceMap()
	if err != nil {
		return err
	}
	resourceSet := make(map[string]struct{}, len(resourceMap.Resources))
	for name := range resourceMap.Resources {
		resourceSet[name] = struct{}{}
	}

	query := strings.TrimSpace(getStringFlag(cmd, "query"))
	keyFilter := strings.TrimSpace(getStringFlag(cmd, "key"))
	pathFilter := strings.TrimSpace(getStringFlag(cmd, "path"))
	paramFilter := strings.TrimSpace(getStringFlag(cmd, "param"))
	terminalSegmentFilter := strings.TrimSpace(getStringFlag(cmd, "terminal-segment"))
	terminalParamFilter := strings.TrimSpace(getStringFlag(cmd, "terminal-param"))
	resourceFilters := parseCSVFilter(strings.TrimSpace(getStringFlag(cmd, "resource")))

	actionRaw := strings.TrimSpace(getStringFlag(cmd, "action"))
	var actionFilter *bool
	if actionRaw != "" {
		parsed, err := strconv.ParseBool(actionRaw)
		if err != nil {
			return fmt.Errorf("--action must be true or false")
		}
		actionFilter = &parsed
	}

	useFullURL := getBoolFlag(cmd, "full-url")
	baseURL := ""
	if useFullURL {
		baseURL = resolveClientBaseURL(cmd)
	}

	rows := []knowledgeClientRouteRow{}
	for _, route := range catalog.Routes {
		binding := buildClientRouteBinding(route, resourceSet)
		resources := binding.ReferencedResources

		if actionFilter != nil && route.Action != *actionFilter {
			continue
		}
		if len(resourceFilters) > 0 && !matchesAnyResource(resources, resourceFilters) {
			continue
		}
		if keyFilter != "" && !containsFold(route.Key, keyFilter) {
			continue
		}
		if pathFilter != "" && !containsFold(route.Path, pathFilter) {
			continue
		}
		if paramFilter != "" && !containsAnyFold(route.Params, paramFilter) {
			continue
		}
		if terminalSegmentFilter != "" && !containsFold(route.TerminalSegment, terminalSegmentFilter) {
			continue
		}
		if terminalParamFilter != "" && !containsFold(route.TerminalParam, terminalParamFilter) {
			continue
		}
		if query != "" && !routeMatchesQuery(route, resources, query) {
			continue
		}

		path := route.Path
		if useFullURL {
			path = clientURL(baseURL, path)
		}

		rows = append(rows, knowledgeClientRouteRow{
			Key:             route.Key,
			Path:            path,
			Params:          append([]string(nil), route.Params...),
			Action:          route.Action,
			TerminalSegment: route.TerminalSegment,
			TerminalParam:   route.TerminalParam,
			Resources:       append([]string(nil), resources...),
		})
	}

	if len(rows) == 0 {
		if len(resourceFilters) > 0 {
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"No client routes found for resource filter %q. Try 'xbe knowledge client-routes --query %s' to search route keys/paths directly.\n",
				strings.Join(resourceFilters, ","),
				resourceFilters[0],
			)
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "No client routes found.")
		return nil
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Path == rows[j].Path {
			return rows[i].Key < rows[j].Key
		}
		return rows[i].Path < rows[j].Path
	})

	offset := getIntFlag(cmd, "offset")
	limit := getIntFlag(cmd, "limit")
	if offset < 0 {
		offset = 0
	}
	if offset >= len(rows) {
		rows = []knowledgeClientRouteRow{}
	} else if offset > 0 {
		rows = rows[offset:]
	}
	if limit <= 0 {
		limit = 200
	}
	if limit < len(rows) {
		rows = rows[:limit]
	}

	if len(rows) == 0 {
		if len(resourceFilters) > 0 {
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"No client routes found for resource filter %q. Try 'xbe knowledge client-routes --query %s' to search route keys/paths directly.\n",
				strings.Join(resourceFilters, ","),
				resourceFilters[0],
			)
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "No client routes found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, rows)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "KEY\tPATH\tPARAMS\tACTION\tTERMINAL\tRESOURCES")
	for _, row := range rows {
		action := ""
		if row.Action {
			action = "yes"
		}
		params := strings.Join(row.Params, ",")
		terminal := formatTerminal(row.TerminalSegment, row.TerminalParam)
		resources := strings.Join(row.Resources, ",")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", row.Key, row.Path, params, action, terminal, resources)
	}
	return w.Flush()
}

func routeMatchesQuery(route clientRoute, resources []string, query string) bool {
	if containsFold(route.Key, query) || containsFold(route.Path, query) {
		return true
	}
	if containsAnyFold(route.Params, query) {
		return true
	}
	if containsFold(route.TerminalSegment, query) || containsFold(route.TerminalParam, query) {
		return true
	}
	return containsAnyFold(resources, query)
}

func containsFold(haystack, needle string) bool {
	haystack = strings.ToLower(strings.TrimSpace(haystack))
	needle = strings.ToLower(strings.TrimSpace(needle))
	if needle == "" {
		return true
	}
	return strings.Contains(haystack, needle)
}

func containsAnyFold(values []string, needle string) bool {
	for _, value := range values {
		if containsFold(value, needle) {
			return true
		}
	}
	return false
}

func matchesAnyResource(resources, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if containsAnyFold(resources, filter) {
			return true
		}
	}
	return false
}

func formatTerminal(segment, param string) string {
	segment = strings.TrimSpace(segment)
	param = strings.TrimSpace(param)
	if segment == "" && param == "" {
		return ""
	}
	if segment == "" {
		return param
	}
	if param == "" {
		return segment
	}
	return segment + ":" + param
}
