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

type doLineupsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	Customer   string
	Name       string
	StartAtMin string
	StartAtMax string
}

func newDoLineupsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup",
		Long: `Create a lineup.

Required flags:
  --customer      Customer ID
  --start-at-min  Earliest start time (ISO 8601)
  --start-at-max  Latest start time (ISO 8601)

Optional flags:
  --name          Lineup name

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a lineup
  xbe do lineups create --customer 123 --start-at-min 2026-01-01T06:00:00Z --start-at-max 2026-01-01T18:00:00Z

  # Create a named lineup
  xbe do lineups create --customer 123 --name "Morning" --start-at-min 2026-01-01T06:00:00Z --start-at-max 2026-01-01T12:00:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoLineupsCreate,
	}
	initDoLineupsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupsCmd.AddCommand(newDoLineupsCreateCmd())
}

func initDoLineupsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID (required)")
	cmd.Flags().String("name", "", "Lineup name")
	cmd.Flags().String("start-at-min", "", "Earliest start time (ISO 8601, required)")
	cmd.Flags().String("start-at-max", "", "Latest start time (ISO 8601, required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("customer")
	_ = cmd.MarkFlagRequired("start-at-min")
	_ = cmd.MarkFlagRequired("start-at-max")
}

func runDoLineupsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Customer) == "" {
		err := fmt.Errorf("--customer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.StartAtMin) == "" {
		err := fmt.Errorf("--start-at-min is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.StartAtMax) == "" {
		err := fmt.Errorf("--start-at-max is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"start-at-min": opts.StartAtMin,
		"start-at-max": opts.StartAtMax,
	}
	setStringAttrIfPresent(attributes, "name", opts.Name)

	relationships := map[string]any{
		"customer": map[string]any{
			"data": map[string]string{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineups",
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

	body, _, err := client.Post(cmd.Context(), "/v1/lineups", jsonBody)
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

	details := buildLineupDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup %s\n", details.ID)
	return nil
}

func parseDoLineupsCreateOptions(cmd *cobra.Command) (doLineupsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customer, _ := cmd.Flags().GetString("customer")
	name, _ := cmd.Flags().GetString("name")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		Customer:   customer,
		Name:       name,
		StartAtMin: startAtMin,
		StartAtMax: startAtMax,
	}, nil
}
