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

type doTruckerBrokeragesCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	Trucker         string
	BrokeredTrucker string
}

func newDoTruckerBrokeragesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trucker brokerage",
		Long: `Create a trucker brokerage.

Required flags:
  --trucker           Brokering trucker ID
  --brokered-trucker  Brokered trucker ID`,
		Example: `  # Create a trucker brokerage
  xbe do trucker-brokerages create --trucker 123 --brokered-trucker 456

  # Get JSON output
  xbe do trucker-brokerages create --trucker 123 --brokered-trucker 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTruckerBrokeragesCreate,
	}
	initDoTruckerBrokeragesCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerBrokeragesCmd.AddCommand(newDoTruckerBrokeragesCreateCmd())
}

func initDoTruckerBrokeragesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Brokering trucker ID (required)")
	cmd.Flags().String("brokered-trucker", "", "Brokered trucker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerBrokeragesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerBrokeragesCreateOptions(cmd)
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

	truckerID := strings.TrimSpace(opts.Trucker)
	if truckerID == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	brokeredTruckerID := strings.TrimSpace(opts.BrokeredTrucker)
	if brokeredTruckerID == "" {
		err := fmt.Errorf("--brokered-trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   truckerID,
			},
		},
		"brokered-trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   brokeredTruckerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trucker-brokerages",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-brokerages", jsonBody)
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

	details := buildTruckerBrokerageDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker brokerage %s\n", details.ID)
	return nil
}

func parseDoTruckerBrokeragesCreateOptions(cmd *cobra.Command) (doTruckerBrokeragesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
	brokeredTrucker, _ := cmd.Flags().GetString("brokered-trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerBrokeragesCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		Trucker:         trucker,
		BrokeredTrucker: brokeredTrucker,
	}, nil
}
