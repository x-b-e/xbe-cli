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

type doJobProductionPlanCancellationReasonTypesCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Name        string
	Slug        string
	Description string
}

func newDoJobProductionPlanCancellationReasonTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cancellation reason type",
		Long: `Create a new job production plan cancellation reason type.

Required flags:
  --name  The type name (required)
  --slug  URL-friendly identifier (required, cannot be changed after creation)

Optional flags:
  --description  Type description`,
		Example: `  # Create a cancellation reason type
  xbe do job-production-plan-cancellation-reason-types create --name "Weather" --slug "weather"

  # Create with description
  xbe do job-production-plan-cancellation-reason-types create --name "Weather" --slug "weather" --description "Cancelled due to weather conditions"

  # Get JSON output
  xbe do job-production-plan-cancellation-reason-types create --name "Equipment" --slug "equipment" --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanCancellationReasonTypesCreate,
	}
	initDoJobProductionPlanCancellationReasonTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanCancellationReasonTypesCmd.AddCommand(newDoJobProductionPlanCancellationReasonTypesCreateCmd())
}

func initDoJobProductionPlanCancellationReasonTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Type name (required)")
	cmd.Flags().String("slug", "", "URL-friendly identifier (required)")
	cmd.Flags().String("description", "", "Type description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanCancellationReasonTypesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanCancellationReasonTypesCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Slug == "" {
		err := fmt.Errorf("--slug is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
		"slug": opts.Slug,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-production-plan-cancellation-reason-types",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-cancellation-reason-types", jsonBody)
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

	row := buildJobProductionPlanCancellationReasonTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created cancellation reason type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoJobProductionPlanCancellationReasonTypesCreateOptions(cmd *cobra.Command) (doJobProductionPlanCancellationReasonTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	slug, _ := cmd.Flags().GetString("slug")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanCancellationReasonTypesCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Name:        name,
		Slug:        slug,
		Description: description,
	}, nil
}

func buildJobProductionPlanCancellationReasonTypeRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanCancellationReasonTypeRow {
	attrs := resp.Data.Attributes

	return jobProductionPlanCancellationReasonTypeRow{
		ID:          resp.Data.ID,
		Slug:        stringAttr(attrs, "slug"),
		Name:        stringAttr(attrs, "name"),
		Description: stringAttr(attrs, "description"),
	}
}
