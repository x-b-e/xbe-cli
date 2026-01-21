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

type doProjectOfficesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Name         string
	Abbreviation string
	IsActive     bool
	Broker       string
}

func newDoProjectOfficesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project office",
		Long: `Create a new project office.

Required flags:
  --name    The project office name (required)
  --broker  The broker ID (required)

Optional flags:
  --abbreviation  Short code for the project office
  --is-active     Whether the office is active (default: true)`,
		Example: `  # Create a project office
  xbe do project-offices create --name "Chicago Office" --broker 123

  # Create with abbreviation
  xbe do project-offices create --name "Chicago Office" --abbreviation "CHI" --broker 123

  # Create inactive
  xbe do project-offices create --name "Old Office" --broker 123 --is-active=false`,
		Args: cobra.NoArgs,
		RunE: runDoProjectOfficesCreate,
	}
	initDoProjectOfficesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectOfficesCmd.AddCommand(newDoProjectOfficesCreateCmd())
}

func initDoProjectOfficesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Project office name (required)")
	cmd.Flags().String("abbreviation", "", "Short code")
	cmd.Flags().Bool("is-active", true, "Whether active")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectOfficesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectOfficesCreateOptions(cmd)
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

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name":      opts.Name,
		"is-active": opts.IsActive,
	}

	if opts.Abbreviation != "" {
		attributes["abbreviation"] = opts.Abbreviation
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-offices",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-offices", jsonBody)
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

	row := buildProjectOfficeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project office %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectOfficesCreateOptions(cmd *cobra.Command) (doProjectOfficesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	isActive, _ := cmd.Flags().GetBool("is-active")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectOfficesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Name:         name,
		Abbreviation: abbreviation,
		IsActive:     isActive,
		Broker:       broker,
	}, nil
}

func buildProjectOfficeRowFromSingle(resp jsonAPISingleResponse) projectOfficeRow {
	attrs := resp.Data.Attributes

	row := projectOfficeRow{
		ID:           resp.Data.ID,
		Name:         stringAttr(attrs, "name"),
		Abbreviation: stringAttr(attrs, "abbreviation"),
		IsActive:     boolAttr(attrs, "is-active"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
