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

type doBrokerTruckerRatingsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	BrokerID  string
	TruckerID string
	Rating    int
}

func newDoBrokerTruckerRatingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker trucker rating",
		Long: `Create a broker trucker rating.

Required flags:
  --broker   Broker ID (required)
  --trucker  Trucker ID (required)
  --rating   Rating (1-5, required)`,
		Example: `  # Create a broker trucker rating
  xbe do broker-trucker-ratings create --broker 123 --trucker 456 --rating 5`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerTruckerRatingsCreate,
	}
	initDoBrokerTruckerRatingsCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerTruckerRatingsCmd.AddCommand(newDoBrokerTruckerRatingsCreateCmd())
}

func initDoBrokerTruckerRatingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().Int("rating", 0, "Rating (1-5, required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerTruckerRatingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerTruckerRatingsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.BrokerID) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.TruckerID) == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if !cmd.Flags().Changed("rating") {
		err := fmt.Errorf("--rating is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"rating": opts.Rating,
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
		"trucker": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.TruckerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "broker-trucker-ratings",
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

	body, _, err := client.Post(cmd.Context(), "/v1/broker-trucker-ratings", jsonBody)
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

	row := brokerTruckerRatingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker trucker rating %s\n", row.ID)
	return nil
}

func parseDoBrokerTruckerRatingsCreateOptions(cmd *cobra.Command) (doBrokerTruckerRatingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerID, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	rating, _ := cmd.Flags().GetInt("rating")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerTruckerRatingsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		BrokerID:  brokerID,
		TruckerID: trucker,
		Rating:    rating,
	}, nil
}
