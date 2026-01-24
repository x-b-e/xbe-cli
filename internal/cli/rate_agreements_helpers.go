package cli

import (
	"fmt"
	"strings"
)

func parseRateAgreementParty(value string, role string) (string, string, error) {
	if strings.TrimSpace(value) == "" {
		return "", "", fmt.Errorf("--%s is required (format: Type|ID, e.g. Broker|123)", role)
	}

	partyType, partyID, err := parseOrganization(value)
	if err != nil {
		return "", "", err
	}

	switch role {
	case "seller":
		if partyType != "brokers" && partyType != "truckers" {
			return "", "", fmt.Errorf("--seller must be Broker|ID or Trucker|ID")
		}
	case "buyer":
		if partyType != "brokers" && partyType != "customers" {
			return "", "", fmt.Errorf("--buyer must be Broker|ID or Customer|ID")
		}
	}

	return partyType, partyID, nil
}
