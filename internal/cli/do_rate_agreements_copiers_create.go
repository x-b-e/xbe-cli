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

type doRateAgreementsCopiersCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	RateAgreementTemplate string
	TargetCustomers       []string
	TargetTruckers        []string
	Note                  string
}

func newDoRateAgreementsCopiersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Copy a rate agreement template",
		Long: `Copy a rate agreement template to customers or truckers.

Required flags:
  --rate-agreement-template  Template rate agreement ID
  --target-customers         Target customer IDs (comma-separated or repeated)
  --target-truckers          Target trucker IDs (comma-separated or repeated)

Optional flags:
  --note  Copier note

Note: Provide exactly one of --target-customers or --target-truckers.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Copy a template to customers
  xbe do rate-agreements-copiers create \
    --rate-agreement-template 123 \
    --target-customers 456,789 \
    --note "Annual renewal"

  # Copy a template to truckers
  xbe do rate-agreements-copiers create \
    --rate-agreement-template 123 \
    --target-truckers 555,666`,
		Args: cobra.NoArgs,
		RunE: runDoRateAgreementsCopiersCreate,
	}
	initDoRateAgreementsCopiersCreateFlags(cmd)
	return cmd
}

func init() {
	doRateAgreementsCopiersCmd.AddCommand(newDoRateAgreementsCopiersCreateCmd())
}

func initDoRateAgreementsCopiersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rate-agreement-template", "", "Template rate agreement ID")
	cmd.Flags().StringSlice("target-customers", nil, "Target customer IDs (comma-separated or repeated)")
	cmd.Flags().StringSlice("target-truckers", nil, "Target trucker IDs (comma-separated or repeated)")
	cmd.Flags().String("note", "", "Copier note")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRateAgreementsCopiersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRateAgreementsCopiersCreateOptions(cmd)
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

	opts.RateAgreementTemplate = strings.TrimSpace(opts.RateAgreementTemplate)
	if opts.RateAgreementTemplate == "" {
		err := fmt.Errorf("--rate-agreement-template is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	targetCustomers := cleanIDList(opts.TargetCustomers)
	targetTruckers := cleanIDList(opts.TargetTruckers)
	if len(targetCustomers) > 0 && len(targetTruckers) > 0 {
		err := fmt.Errorf("only one of --target-customers or --target-truckers can be set")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if len(targetCustomers) == 0 && len(targetTruckers) == 0 {
		err := fmt.Errorf("one of --target-customers or --target-truckers is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	setStringAttrIfPresent(attributes, "note", opts.Note)

	relationships := map[string]any{
		"rate-agreement-template": map[string]any{
			"data": map[string]any{
				"type": "rate-agreements",
				"id":   opts.RateAgreementTemplate,
			},
		},
	}

	if len(targetCustomers) > 0 {
		relationships["target-customers"] = map[string]any{
			"data": buildRelationshipDataList("customers", targetCustomers),
		}
	}
	if len(targetTruckers) > 0 {
		relationships["target-truckers"] = map[string]any{
			"data": buildRelationshipDataList("truckers", targetTruckers),
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "rate-agreements-copiers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/rate-agreements-copiers", jsonBody)
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

	if opts.JSON {
		rows := buildRateAgreementsCopierRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
		if len(rows) > 0 {
			return writeJSON(cmd.OutOrStdout(), rows[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created rate agreements copier %s\n", resp.Data.ID)
	return nil
}

func parseDoRateAgreementsCopiersCreateOptions(cmd *cobra.Command) (doRateAgreementsCopiersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rateAgreementTemplate, _ := cmd.Flags().GetString("rate-agreement-template")
	targetCustomers, _ := cmd.Flags().GetStringSlice("target-customers")
	targetTruckers, _ := cmd.Flags().GetStringSlice("target-truckers")
	note, _ := cmd.Flags().GetString("note")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRateAgreementsCopiersCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		RateAgreementTemplate: rateAgreementTemplate,
		TargetCustomers:       targetCustomers,
		TargetTruckers:        targetTruckers,
		Note:                  note,
	}, nil
}

func cleanIDList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}
