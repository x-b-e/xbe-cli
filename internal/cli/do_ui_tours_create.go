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

type doUiToursCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Name         string
	Abbreviation string
	Description  string
}

func newDoUiToursCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new UI tour",
		Long: `Create a new UI tour.

Required flags:
  --name          UI tour name (required)
  --abbreviation  UI tour abbreviation (required)

Optional flags:
  --description   UI tour description

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a UI tour
  xbe do ui-tours create --name "Driver Onboarding" --abbreviation "driver-onboarding"

  # Create with description
  xbe do ui-tours create --name "Project Setup" --abbreviation "project-setup" --description "Walkthrough for project creation"

  # Get JSON output
  xbe do ui-tours create --name "QA" --abbreviation "qa" --json`,
		Args: cobra.NoArgs,
		RunE: runDoUiToursCreate,
	}
	initDoUiToursCreateFlags(cmd)
	return cmd
}

func init() {
	doUiToursCmd.AddCommand(newDoUiToursCreateCmd())
}

func initDoUiToursCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "UI tour name (required)")
	cmd.Flags().String("abbreviation", "", "UI tour abbreviation (required)")
	cmd.Flags().String("description", "", "UI tour description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUiToursCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoUiToursCreateOptions(cmd)
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

	name := strings.TrimSpace(opts.Name)
	abbreviation := strings.TrimSpace(opts.Abbreviation)
	if name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if abbreviation == "" {
		err := fmt.Errorf("--abbreviation is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name":         name,
		"abbreviation": abbreviation,
	}
	if strings.TrimSpace(opts.Description) != "" {
		attributes["description"] = strings.TrimSpace(opts.Description)
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "ui-tours",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/ui-tours", jsonBody)
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

	row := buildUiTourRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created UI tour %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoUiToursCreateOptions(cmd *cobra.Command) (doUiToursCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUiToursCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Name:         name,
		Abbreviation: abbreviation,
		Description:  description,
	}, nil
}

func buildUiTourRowFromSingle(resp jsonAPISingleResponse) uiTourRow {
	attrs := resp.Data.Attributes

	return uiTourRow{
		ID:           resp.Data.ID,
		Name:         stringAttr(attrs, "name"),
		Abbreviation: stringAttr(attrs, "abbreviation"),
		Description:  stringAttr(attrs, "description"),
	}
}
