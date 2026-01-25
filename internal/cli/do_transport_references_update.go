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

type doTransportReferencesUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	Key      string
	Value    string
	Position int
}

func newDoTransportReferencesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a transport reference",
		Long: `Update a transport reference.

Optional:
  --key       Reference key
  --value     Reference value
  --position  Reference position within the subject`,
		Example: `  # Update key and value
  xbe do transport-references update 123 --key BOL --value "BOL-1000"

  # Update position
  xbe do transport-references update 123 --position 2`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTransportReferencesUpdate,
	}
	initDoTransportReferencesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTransportReferencesCmd.AddCommand(newDoTransportReferencesUpdateCmd())
}

func initDoTransportReferencesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("key", "", "Reference key")
	cmd.Flags().String("value", "", "Reference value")
	cmd.Flags().Int("position", 0, "Reference position within the subject")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportReferencesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportReferencesUpdateOptions(cmd, args)
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

	attributes := map[string]any{}

	if cmd.Flags().Changed("key") {
		attributes["key"] = opts.Key
	}
	if cmd.Flags().Changed("value") {
		attributes["value"] = opts.Value
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "transport-references",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/transport-references/"+opts.ID, jsonBody)
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
		row := buildTransportReferenceRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated transport reference %s\n", resp.Data.ID)
	return nil
}

func parseDoTransportReferencesUpdateOptions(cmd *cobra.Command, args []string) (doTransportReferencesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	key, _ := cmd.Flags().GetString("key")
	value, _ := cmd.Flags().GetString("value")
	position, _ := cmd.Flags().GetInt("position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTransportReferencesUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		Key:      key,
		Value:    value,
		Position: position,
	}, nil
}
