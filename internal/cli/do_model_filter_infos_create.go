package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doModelFilterInfosCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ResourceType     string
	FilterKeysRaw    string
	FilterKeys       []string
	ScopeFiltersJSON string
	ScopeFilterPairs []string
}

type modelFilterInfoDetails struct {
	ID           string         `json:"id,omitempty"`
	ResourceType string         `json:"resource_type,omitempty"`
	FilterKeys   []string       `json:"filter_keys,omitempty"`
	ScopeFilters map[string]any `json:"scope_filters,omitempty"`
	Options      any            `json:"options,omitempty"`
}

func newDoModelFilterInfosCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Fetch filter options for a resource",
		Long: `Fetch filter options for a resource.

Provide a resource type and optionally limit filter keys or scope the option
query using filters.

Required flags:
  --resource-type  Resource type (e.g. projects, users) (required)

Optional flags:
  --filter-keys    Filter keys to include (comma-separated)
  --filter-key     Filter key to include (repeatable)
  --scope-filters  Scope filters JSON object
  --scope-filter   Scope filter in key=value format (repeatable)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Fetch filter options for projects
  xbe do model-filter-infos create --resource-type projects

  # Limit to selected filter keys
  xbe do model-filter-infos create --resource-type projects --filter-keys customer,project_manager

  # Scope options to a broker
  xbe do model-filter-infos create --resource-type projects --scope-filter broker=123

  # JSON output
  xbe do model-filter-infos create --resource-type projects --json`,
		Args: cobra.NoArgs,
		RunE: runDoModelFilterInfosCreate,
	}
	initDoModelFilterInfosCreateFlags(cmd)
	return cmd
}

func init() {
	doModelFilterInfosCmd.AddCommand(newDoModelFilterInfosCreateCmd())
}

func initDoModelFilterInfosCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("resource-type", "", "Resource type (required)")
	cmd.Flags().String("filter-keys", "", "Filter keys to include (comma-separated)")
	cmd.Flags().StringArray("filter-key", nil, "Filter key to include (repeatable)")
	cmd.Flags().String("scope-filters", "", "Scope filters JSON object")
	cmd.Flags().StringArray("scope-filter", nil, "Scope filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("resource-type")
}

func runDoModelFilterInfosCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoModelFilterInfosCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	resourceType := strings.TrimSpace(opts.ResourceType)
	if resourceType == "" {
		err := errors.New("--resource-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	scopeFilters, err := parseModelFilterInfoScopeFilters(opts.ScopeFiltersJSON, opts.ScopeFilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	filterKeys := resolveModelFilterInfoFilterKeys(opts)

	attributes := map[string]any{
		"resource-type": resourceType,
	}

	if len(filterKeys) > 0 {
		attributes["filter-keys"] = filterKeys
	}
	if len(scopeFilters) > 0 {
		attributes["scope-filters"] = scopeFilters
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "model-filter-infos",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/model-filter-infos", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildModelFilterInfoDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderModelFilterInfoDetails(cmd, details)
}

func parseDoModelFilterInfosCreateOptions(cmd *cobra.Command) (doModelFilterInfosCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}
	resourceType, err := cmd.Flags().GetString("resource-type")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}
	filterKeysRaw, err := cmd.Flags().GetString("filter-keys")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}
	filterKeys, err := cmd.Flags().GetStringArray("filter-key")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}
	scopeFiltersJSON, err := cmd.Flags().GetString("scope-filters")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}
	scopeFilterPairs, err := cmd.Flags().GetStringArray("scope-filter")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doModelFilterInfosCreateOptions{}, err
	}

	return doModelFilterInfosCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ResourceType:     resourceType,
		FilterKeysRaw:    filterKeysRaw,
		FilterKeys:       filterKeys,
		ScopeFiltersJSON: scopeFiltersJSON,
		ScopeFilterPairs: scopeFilterPairs,
	}, nil
}

func resolveModelFilterInfoFilterKeys(opts doModelFilterInfosCreateOptions) []string {
	keys := []string{}
	if opts.FilterKeysRaw != "" {
		keys = append(keys, splitCommaList(opts.FilterKeysRaw)...)
	}
	if len(opts.FilterKeys) > 0 {
		keys = append(keys, opts.FilterKeys...)
	}
	return uniqueStrings(keys)
}

func parseModelFilterInfoScopeFilters(rawJSON string, pairs []string) (map[string]any, error) {
	filters := map[string]any{}

	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON != "" {
		if err := json.Unmarshal([]byte(rawJSON), &filters); err != nil {
			return nil, fmt.Errorf("invalid --scope-filters JSON: %w", err)
		}
	}

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --scope-filter %q (expected key=value)", pair)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("invalid --scope-filter %q (missing key)", pair)
		}
		filters[key] = value
	}

	return filters, nil
}

func buildModelFilterInfoDetails(resp jsonAPISingleResponse) modelFilterInfoDetails {
	attrs := resp.Data.Attributes
	details := modelFilterInfoDetails{
		ID:           resp.Data.ID,
		ResourceType: stringAttr(attrs, "resource-type"),
		FilterKeys:   stringSliceAttr(attrs, "filter-keys"),
		ScopeFilters: mapAttr(attrs, "scope-filters"),
		Options:      anyAttr(attrs, "options"),
	}

	return details
}

func mapAttr(attrs map[string]any, key string) map[string]any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case map[string]any:
		return typed
	case map[string]string:
		out := make(map[string]any, len(typed))
		for k, v := range typed {
			out[k] = v
		}
		return out
	default:
		return nil
	}
}

func renderModelFilterInfoDetails(cmd *cobra.Command, details modelFilterInfoDetails) error {
	out := cmd.OutOrStdout()

	if details.ID != "" {
		fmt.Fprintf(out, "ID: %s\n", details.ID)
	}
	if details.ResourceType != "" {
		fmt.Fprintf(out, "Resource Type: %s\n", details.ResourceType)
	}
	if len(details.FilterKeys) > 0 {
		fmt.Fprintf(out, "Filter Keys: %s\n", strings.Join(details.FilterKeys, ", "))
	}
	if len(details.ScopeFilters) > 0 {
		fmt.Fprintf(out, "Scope Filters: %s\n", formatAny(details.ScopeFilters))
	}
	if details.Options != nil {
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, formatAny(details.Options))
	}

	return nil
}
