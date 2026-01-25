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

type doSourcingSearchesCreateOptions struct {
	BaseURL                                   string
	Token                                     string
	JSON                                      bool
	CustomerTenderID                          string
	MaximumDistanceMiles                      float64
	MaximumResultCount                        int
	MinimumTruckerRating                      float64
	AdditionalCertificationRequirementTypeIDs []string
}

type sourcingSearchRow struct {
	ID                                        string   `json:"id"`
	CustomerTenderID                          string   `json:"customer_tender_id,omitempty"`
	MaximumDistanceMiles                      float64  `json:"maximum_distance_miles,omitempty"`
	MaximumResultCount                        int      `json:"maximum_result_count,omitempty"`
	MinimumTruckerRating                      *float64 `json:"minimum_trucker_rating,omitempty"`
	AdditionalCertificationRequirementTypeIDs []string `json:"additional_certification_requirement_type_ids,omitempty"`
	ResultIDs                                 []string `json:"result_ids,omitempty"`
	ResultCount                               int      `json:"result_count,omitempty"`
}

func newDoSourcingSearchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Run a sourcing search",
		Long: `Run a sourcing search to find matching truckers, trailers, and broker tenders.

Required flags:
  --customer-tender   Customer tender ID

Optional flags:
  --maximum-distance-miles                 Maximum distance in miles
  --maximum-result-count                   Maximum number of results to return
  --minimum-trucker-rating                 Minimum trucker rating
  --additional-certification-requirement-types Additional certification type IDs (comma-separated or repeated)`,
		Example: `  # Run a sourcing search with defaults
  xbe do sourcing-searches create --customer-tender 123

  # Constrain the search
  xbe do sourcing-searches create --customer-tender 123 \
    --maximum-distance-miles 75 --maximum-result-count 25

  # Require additional certification types
  xbe do sourcing-searches create --customer-tender 123 \
    --additional-certification-requirement-types 45,67

  # JSON output
  xbe do sourcing-searches create --customer-tender 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoSourcingSearchesCreate,
	}
	initDoSourcingSearchesCreateFlags(cmd)
	return cmd
}

func init() {
	doSourcingSearchesCmd.AddCommand(newDoSourcingSearchesCreateCmd())
}

func initDoSourcingSearchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer-tender", "", "Customer tender ID (required)")
	cmd.Flags().Float64("maximum-distance-miles", 0, "Maximum distance in miles")
	cmd.Flags().Int("maximum-result-count", 0, "Maximum number of results to return")
	cmd.Flags().Float64("minimum-trucker-rating", 0, "Minimum trucker rating")
	cmd.Flags().StringSlice("additional-certification-requirement-types", nil, "Additional certification type IDs (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("customer-tender")
}

func runDoSourcingSearchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoSourcingSearchesCreateOptions(cmd)
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

	customerTenderID := strings.TrimSpace(opts.CustomerTenderID)
	if customerTenderID == "" {
		err := fmt.Errorf("--customer-tender is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("maximum-distance-miles") {
		attributes["maximum-distance-miles"] = opts.MaximumDistanceMiles
	}
	if cmd.Flags().Changed("maximum-result-count") {
		attributes["maximum-result-count"] = opts.MaximumResultCount
	}
	if cmd.Flags().Changed("minimum-trucker-rating") {
		attributes["minimum-trucker-rating"] = opts.MinimumTruckerRating
	}

	relationships := map[string]any{
		"customer-tender": map[string]any{
			"data": map[string]any{
				"type": "customer-tenders",
				"id":   customerTenderID,
			},
		},
	}

	additionalTypes := compactStringSlice(opts.AdditionalCertificationRequirementTypeIDs)
	if len(additionalTypes) > 0 {
		relationships["additional-certification-requirement-types"] = map[string]any{
			"data": buildRelationshipData("certification-types", additionalTypes),
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "sourcing-searches",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/sourcing-searches", jsonBody)
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

	row := buildSourcingSearchRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderSourcingSearch(cmd, row)
}

func parseDoSourcingSearchesCreateOptions(cmd *cobra.Command) (doSourcingSearchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customerTenderID, _ := cmd.Flags().GetString("customer-tender")
	maximumDistanceMiles, _ := cmd.Flags().GetFloat64("maximum-distance-miles")
	maximumResultCount, _ := cmd.Flags().GetInt("maximum-result-count")
	minimumTruckerRating, _ := cmd.Flags().GetFloat64("minimum-trucker-rating")
	additionalTypes, _ := cmd.Flags().GetStringSlice("additional-certification-requirement-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSourcingSearchesCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		CustomerTenderID:     customerTenderID,
		MaximumDistanceMiles: maximumDistanceMiles,
		MaximumResultCount:   maximumResultCount,
		MinimumTruckerRating: minimumTruckerRating,
		AdditionalCertificationRequirementTypeIDs: additionalTypes,
	}, nil
}

func buildSourcingSearchRowFromSingle(resp jsonAPISingleResponse) sourcingSearchRow {
	resource := resp.Data
	row := sourcingSearchRow{
		ID:                   resource.ID,
		CustomerTenderID:     relationshipIDFromMap(resource.Relationships, "customer-tender"),
		MaximumDistanceMiles: floatAttr(resource.Attributes, "maximum-distance-miles"),
		MaximumResultCount:   intAttr(resource.Attributes, "maximum-result-count"),
		AdditionalCertificationRequirementTypeIDs: relationshipIDsFromMap(resource.Relationships, "additional-certification-requirement-types"),
		ResultIDs: relationshipIDsFromMap(resource.Relationships, "results"),
	}

	row.MinimumTruckerRating = floatAttrPointer(resource.Attributes, "minimum-trucker-rating")
	row.ResultCount = len(row.ResultIDs)

	return row
}

func renderSourcingSearch(cmd *cobra.Command, row sourcingSearchRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Sourcing search %s\n", row.ID)
	if row.CustomerTenderID != "" {
		fmt.Fprintf(out, "Customer tender: %s\n", row.CustomerTenderID)
	}
	if row.MaximumDistanceMiles > 0 {
		fmt.Fprintf(out, "Maximum distance (miles): %.2f\n", row.MaximumDistanceMiles)
	}
	if row.MaximumResultCount > 0 {
		fmt.Fprintf(out, "Maximum results: %d\n", row.MaximumResultCount)
	}
	if row.MinimumTruckerRating != nil {
		fmt.Fprintf(out, "Minimum trucker rating: %s\n", formatFloat(row.MinimumTruckerRating, 2))
	}
	if len(row.AdditionalCertificationRequirementTypeIDs) > 0 {
		fmt.Fprintf(out, "Additional certification requirement types (%d):\n", len(row.AdditionalCertificationRequirementTypeIDs))
		for _, typeID := range row.AdditionalCertificationRequirementTypeIDs {
			fmt.Fprintf(out, "  %s\n", typeID)
		}
	}

	if len(row.ResultIDs) == 0 {
		fmt.Fprintln(out, "Results: none")
		return nil
	}

	fmt.Fprintf(out, "Results (%d):\n", len(row.ResultIDs))
	for _, resultID := range row.ResultIDs {
		fmt.Fprintf(out, "  %s\n", resultID)
	}

	return nil
}
