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

type doLaborersUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	IsActive           bool
	MobilizationMethod string
	GroupName          string
	ColorHex           string
	CraftClassID       string
}

func newDoLaborersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a laborer",
		Long: `Update a laborer.

Note: labor-classification, user, and organization cannot be changed after creation.

Optional flags:
  --is-active              Whether laborer is active
  --mobilization-method    Mobilization method
  --group-name             Group name
  --color-hex              Color hex code
  --craft-class            Craft class ID`,
		Example: `  # Update active status
  xbe do laborers update 123 --is-active false

  # Update group name
  xbe do laborers update 123 --group-name "Crew B"

  # Update craft class
  xbe do laborers update 123 --craft-class 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLaborersUpdate,
	}
	initDoLaborersUpdateFlags(cmd)
	return cmd
}

func init() {
	doLaborersCmd.AddCommand(newDoLaborersUpdateCmd())
}

func initDoLaborersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-active", false, "Whether laborer is active")
	cmd.Flags().String("mobilization-method", "", "Mobilization method")
	cmd.Flags().String("group-name", "", "Group name")
	cmd.Flags().String("color-hex", "", "Color hex code")
	cmd.Flags().String("craft-class", "", "Craft class ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLaborersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLaborersUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}
	if cmd.Flags().Changed("mobilization-method") {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}
	if cmd.Flags().Changed("group-name") {
		attributes["group-name"] = opts.GroupName
	}
	if cmd.Flags().Changed("color-hex") {
		attributes["color-hex"] = opts.ColorHex
	}

	if cmd.Flags().Changed("craft-class") {
		if opts.CraftClassID == "" {
			relationships["craft-class"] = map[string]any{"data": nil}
		} else {
			relationships["craft-class"] = map[string]any{
				"data": map[string]any{
					"type": "craft-classes",
					"id":   opts.CraftClassID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "laborers",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/laborers/"+opts.ID, jsonBody)
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

	row := buildLaborerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated laborer %s\n", row.ID)
	return nil
}

func parseDoLaborersUpdateOptions(cmd *cobra.Command, args []string) (doLaborersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isActive, _ := cmd.Flags().GetBool("is-active")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	groupName, _ := cmd.Flags().GetString("group-name")
	colorHex, _ := cmd.Flags().GetString("color-hex")
	craftClassID, _ := cmd.Flags().GetString("craft-class")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLaborersUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		IsActive:           isActive,
		MobilizationMethod: mobilizationMethod,
		GroupName:          groupName,
		ColorHex:           colorHex,
		CraftClassID:       craftClassID,
	}, nil
}
