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

type doCraftsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Name    string
	Code    string
	Broker  string
}

func newDoCraftsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new craft",
		Long: `Create a new craft.

Required flags:
  --name    The craft name (required)
  --broker  The broker ID (required)

Optional flags:
  --code    Short code for the craft`,
		Example: `  # Create a basic craft
  xbe do crafts create --name "Carpenter" --broker 123

  # Create with code
  xbe do crafts create --name "Electrician" --code "ELEC" --broker 123

  # Get JSON output
  xbe do crafts create --name "Plumber" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCraftsCreate,
	}
	initDoCraftsCreateFlags(cmd)
	return cmd
}

func init() {
	doCraftsCmd.AddCommand(newDoCraftsCreateCmd())
}

func initDoCraftsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Craft name (required)")
	cmd.Flags().String("code", "", "Short code")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCraftsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCraftsCreateOptions(cmd)
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
		"name": opts.Name,
	}
	if opts.Code != "" {
		attributes["code"] = opts.Code
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "crafts",
			"attributes": attributes,
			"relationships": map[string]any{
				"broker": map[string]any{
					"data": map[string]any{
						"type": "brokers",
						"id":   opts.Broker,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/crafts", jsonBody)
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

	row := buildCraftRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created craft %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCraftsCreateOptions(cmd *cobra.Command) (doCraftsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	code, _ := cmd.Flags().GetString("code")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCraftsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Name:    name,
		Code:    code,
		Broker:  broker,
	}, nil
}

func buildCraftRowFromSingle(resp jsonAPISingleResponse) craftRow {
	attrs := resp.Data.Attributes

	row := craftRow{
		ID:   resp.Data.ID,
		Name: stringAttr(attrs, "name"),
		Code: stringAttr(attrs, "code"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
