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

type doBaseSummaryTemplatesCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	Label           string
	GroupBys        []string
	ExplicitMetrics []string
	FiltersJSON     string
	FilterPairs     []string
	StartDate       string
	EndDate         string
	Broker          string
	CreatedBy       string
}

func newDoBaseSummaryTemplatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a base summary template",
		Long: `Create a base summary template.

Required flags:
  --label   Template label (required)

Optional flags:
  --group-bys         Group-by fields (comma-separated or repeatable)
  --explicit-metrics  Explicit metric list (comma-separated or repeatable)
  --filters           Filters JSON object (e.g. '{"broker":"123"}')
  --filter            Filter in key=value format (repeatable)
  --start-date        Start date (YYYY-MM-DD)
  --end-date          End date (YYYY-MM-DD)
  --broker            Broker ID (scopes template to broker)
  --created-by        Creator user ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a base summary template with filters
  xbe do base-summary-templates create \
    --label "Weekly Summary" \
    --group-bys broker,customer \
    --explicit-metrics count,total_cost \
    --filters '{"broker":"123"}' \
    --start-date 2025-01-01 \
    --end-date 2025-01-31 \
    --broker 123

  # JSON output
  xbe do base-summary-templates create \
    --label "Daily Summary" \
    --filter broker=123 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoBaseSummaryTemplatesCreate,
	}
	initDoBaseSummaryTemplatesCreateFlags(cmd)
	return cmd
}

func init() {
	doBaseSummaryTemplatesCmd.AddCommand(newDoBaseSummaryTemplatesCreateCmd())
}

func initDoBaseSummaryTemplatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("label", "", "Template label (required)")
	cmd.Flags().StringSlice("group-bys", nil, "Group-by fields (comma-separated or repeatable)")
	cmd.Flags().StringSlice("explicit-metrics", nil, "Explicit metrics (comma-separated or repeatable)")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBaseSummaryTemplatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBaseSummaryTemplatesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.Label) == "" {
		err := fmt.Errorf("--label is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	filters, err := parseBaseSummaryTemplateFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"label": opts.Label,
	}
	if len(opts.GroupBys) > 0 {
		attributes["group-bys"] = opts.GroupBys
	}
	if len(opts.ExplicitMetrics) > 0 {
		attributes["explicit-metrics"] = opts.ExplicitMetrics
	}
	if strings.TrimSpace(opts.FiltersJSON) != "" || len(opts.FilterPairs) > 0 {
		attributes["filters"] = filters
	}
	if strings.TrimSpace(opts.StartDate) != "" {
		attributes["start-date"] = opts.StartDate
	}
	if strings.TrimSpace(opts.EndDate) != "" {
		attributes["end-date"] = opts.EndDate
	}

	relationships := map[string]any{}
	if strings.TrimSpace(opts.Broker) != "" {
		relationships["broker"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if strings.TrimSpace(opts.CreatedBy) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "base-summary-templates",
			"attributes": attributes,
		},
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/base-summary-templates", jsonBody)
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

	details := buildBaseSummaryTemplateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	if details.Label != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created base summary template %s (%s)\n", details.ID, details.Label)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created base summary template %s\n", details.ID)
	return nil
}

func parseDoBaseSummaryTemplatesCreateOptions(cmd *cobra.Command) (doBaseSummaryTemplatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	label, _ := cmd.Flags().GetString("label")
	groupBys, _ := cmd.Flags().GetStringSlice("group-bys")
	explicitMetrics, _ := cmd.Flags().GetStringSlice("explicit-metrics")
	filtersJSON, _ := cmd.Flags().GetString("filters")
	filterPairs, _ := cmd.Flags().GetStringArray("filter")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBaseSummaryTemplatesCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		Label:           label,
		GroupBys:        groupBys,
		ExplicitMetrics: explicitMetrics,
		FiltersJSON:     filtersJSON,
		FilterPairs:     filterPairs,
		StartDate:       startDate,
		EndDate:         endDate,
		Broker:          broker,
		CreatedBy:       createdBy,
	}, nil
}

func parseBaseSummaryTemplateFilters(rawJSON string, pairs []string) (map[string]any, error) {
	filters := map[string]any{}

	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON != "" {
		if err := json.Unmarshal([]byte(rawJSON), &filters); err != nil {
			return nil, fmt.Errorf("invalid --filters JSON: %w", err)
		}
	}

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --filter %q (expected key=value)", pair)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("invalid --filter %q (missing key)", pair)
		}
		filters[key] = value
	}

	return filters, nil
}
