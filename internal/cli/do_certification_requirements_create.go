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

type doCertificationRequirementsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	CertificationTypeID string
	RequiredByType      string
	RequiredByID        string
	PeriodStart         string
	PeriodEnd           string
}

func newDoCertificationRequirementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new certification requirement",
		Long: `Create a new certification requirement.

Required flags:
  --certification-type    Certification type ID (required)
  --required-by-type      Type of requiring entity (e.g., projects) (required)
  --required-by-id        ID of requiring entity (required)

Optional flags:
  --period-start    Period start date (YYYY-MM-DD)
  --period-end      Period end date (YYYY-MM-DD)`,
		Example: `  # Create a certification requirement for a project
  xbe do certification-requirements create --certification-type 123 --required-by-type projects --required-by-id 456

  # Create with period dates
  xbe do certification-requirements create --certification-type 123 --required-by-type projects --required-by-id 456 --period-start 2024-01-01 --period-end 2025-01-01`,
		Args: cobra.NoArgs,
		RunE: runDoCertificationRequirementsCreate,
	}
	initDoCertificationRequirementsCreateFlags(cmd)
	return cmd
}

func init() {
	doCertificationRequirementsCmd.AddCommand(newDoCertificationRequirementsCreateCmd())
}

func initDoCertificationRequirementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("certification-type", "", "Certification type ID (required)")
	cmd.Flags().String("required-by-type", "", "Type of requiring entity (e.g., projects) (required)")
	cmd.Flags().String("required-by-id", "", "ID of requiring entity (required)")
	cmd.Flags().String("period-start", "", "Period start date (YYYY-MM-DD)")
	cmd.Flags().String("period-end", "", "Period end date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCertificationRequirementsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCertificationRequirementsCreateOptions(cmd)
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

	if opts.CertificationTypeID == "" {
		err := fmt.Errorf("--certification-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.RequiredByType == "" {
		err := fmt.Errorf("--required-by-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.RequiredByID == "" {
		err := fmt.Errorf("--required-by-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.PeriodStart != "" {
		attributes["period-start"] = opts.PeriodStart
	}
	if opts.PeriodEnd != "" {
		attributes["period-end"] = opts.PeriodEnd
	}

	relationships := map[string]any{
		"certification-type": map[string]any{
			"data": map[string]any{
				"type": "certification-types",
				"id":   opts.CertificationTypeID,
			},
		},
		"required-by": map[string]any{
			"data": map[string]any{
				"type": opts.RequiredByType,
				"id":   opts.RequiredByID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "certification-requirements",
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

	body, _, err := client.Post(cmd.Context(), "/v1/certification-requirements", jsonBody)
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

	row := buildCertificationRequirementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created certification requirement %s\n", row.ID)
	return nil
}

func parseDoCertificationRequirementsCreateOptions(cmd *cobra.Command) (doCertificationRequirementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	certificationTypeID, _ := cmd.Flags().GetString("certification-type")
	requiredByType, _ := cmd.Flags().GetString("required-by-type")
	requiredByID, _ := cmd.Flags().GetString("required-by-id")
	periodStart, _ := cmd.Flags().GetString("period-start")
	periodEnd, _ := cmd.Flags().GetString("period-end")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCertificationRequirementsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		CertificationTypeID: certificationTypeID,
		RequiredByType:      requiredByType,
		RequiredByID:        requiredByID,
		PeriodStart:         periodStart,
		PeriodEnd:           periodEnd,
	}, nil
}

func buildCertificationRequirementRowFromSingle(resp jsonAPISingleResponse) certificationRequirementRow {
	attrs := resp.Data.Attributes

	row := certificationRequirementRow{
		ID:          resp.Data.ID,
		PeriodStart: stringAttr(attrs, "period-start"),
		PeriodEnd:   stringAttr(attrs, "period-end"),
	}

	if rel, ok := resp.Data.Relationships["required-by"]; ok && rel.Data != nil {
		row.RequiredByType = rel.Data.Type
		row.RequiredByID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["certification-type"]; ok && rel.Data != nil {
		row.CertificationTypeID = rel.Data.ID
	}

	return row
}
