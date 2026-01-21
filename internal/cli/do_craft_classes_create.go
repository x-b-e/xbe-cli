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

type doCraftClassesCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	Name              string
	Code              string
	IsValidForDrivers bool
	Craft             string
}

func newDoCraftClassesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new craft class",
		Long: `Create a new craft class.

Required flags:
  --name   The craft class name (required)
  --craft  The parent craft ID (required)

Optional flags:
  --code                  Craft class code
  --is-valid-for-drivers  Whether valid for drivers (default: false)`,
		Example: `  # Create a craft class
  xbe do craft-classes create --name "Journeyman" --craft 123

  # Create with code
  xbe do craft-classes create --name "Journeyman" --code "JRN" --craft 123

  # Create valid for drivers
  xbe do craft-classes create --name "CDL Driver" --craft 123 --is-valid-for-drivers`,
		Args: cobra.NoArgs,
		RunE: runDoCraftClassesCreate,
	}
	initDoCraftClassesCreateFlags(cmd)
	return cmd
}

func init() {
	doCraftClassesCmd.AddCommand(newDoCraftClassesCreateCmd())
}

func initDoCraftClassesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Craft class name (required)")
	cmd.Flags().String("code", "", "Craft class code")
	cmd.Flags().Bool("is-valid-for-drivers", false, "Whether valid for drivers")
	cmd.Flags().String("craft", "", "Parent craft ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCraftClassesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCraftClassesCreateOptions(cmd)
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

	if opts.Craft == "" {
		err := fmt.Errorf("--craft is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name":                 opts.Name,
		"is-valid-for-drivers": opts.IsValidForDrivers,
	}

	if opts.Code != "" {
		attributes["code"] = opts.Code
	}

	relationships := map[string]any{
		"craft": map[string]any{
			"data": map[string]any{
				"type": "crafts",
				"id":   opts.Craft,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "craft-classes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/craft-classes", jsonBody)
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

	row := buildCraftClassRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created craft class %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCraftClassesCreateOptions(cmd *cobra.Command) (doCraftClassesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	code, _ := cmd.Flags().GetString("code")
	isValidForDrivers, _ := cmd.Flags().GetBool("is-valid-for-drivers")
	craft, _ := cmd.Flags().GetString("craft")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCraftClassesCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		Name:              name,
		Code:              code,
		IsValidForDrivers: isValidForDrivers,
		Craft:             craft,
	}, nil
}

func buildCraftClassRowFromSingle(resp jsonAPISingleResponse) craftClassRow {
	attrs := resp.Data.Attributes

	row := craftClassRow{
		ID:                resp.Data.ID,
		Name:              stringAttr(attrs, "name"),
		Code:              stringAttr(attrs, "code"),
		IsValidForDrivers: boolAttr(attrs, "is-valid-for-drivers"),
	}

	if rel, ok := resp.Data.Relationships["craft"]; ok && rel.Data != nil {
		row.CraftID = rel.Data.ID
	}

	return row
}
