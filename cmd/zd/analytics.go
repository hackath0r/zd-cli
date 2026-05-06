package main

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// `zd analytics` wraps the v2 analytics POST endpoints. They take a
// filter body so we expose --body / --data, plus convenience flags
// for the most common filter fields (from / to dates).

func newAnalyticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "Read account analytics (incident / service / team / user)",
	}
	cmd.AddCommand(
		newAnalyticsIncidentsCmd(),
		newAnalyticsServicesCmd(),
		newAnalyticsTeamsCmd(),
		newAnalyticsUsersCmd(),
	)
	return cmd
}

// analyticsBody resolves the filter body for an analytics POST call.
// Honours --body / --data if provided; otherwise builds a minimal
// {"from_date": "...", "to_date": "..."} object from the convenience
// flags. At least one of the inputs must be set.
func analyticsBody(cmd *cobra.Command, fromDate, toDate string) (json.RawMessage, error) {
	body, err := readJSONBody(cmd, "")
	if err != nil {
		return nil, err
	}
	merged, err := mergeBody(body, map[string]any{
		"from_date": nonEmpty(fromDate),
		"to_date":   nonEmpty(toDate),
	})
	if err != nil {
		return nil, err
	}
	return merged, nil
}

func newAnalyticsIncidentsCmd() *cobra.Command {
	var fromDate, toDate string
	cmd := &cobra.Command{
		Use:   "incidents",
		Short: "Incident analytics (POST /api/v2/account/analytics/incident-stats/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := analyticsBody(cmd, fromDate, toDate)
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, "incident analytics", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiV2AccountAnalyticsIncidentStatsWithBodyWithResponse(ctx, jsonContentType, bodyReader(body))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&fromDate, "from", "", "from_date YYYY-MM-DD (convenience flag)")
	cmd.Flags().StringVar(&toDate, "to", "", "to_date YYYY-MM-DD (convenience flag)")
	return cmd
}

func newAnalyticsServicesCmd() *cobra.Command {
	var fromDate, toDate string
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Service analytics (POST /api/v2/account/analytics/service-analytics/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := analyticsBody(cmd, fromDate, toDate)
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, "service analytics", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiV2AccountAnalyticsServiceAnalyticsWithBodyWithResponse(ctx, jsonContentType, bodyReader(body))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&fromDate, "from", "", "from_date YYYY-MM-DD (convenience flag)")
	cmd.Flags().StringVar(&toDate, "to", "", "to_date YYYY-MM-DD (convenience flag)")
	return cmd
}

func newAnalyticsTeamsCmd() *cobra.Command {
	var fromDate, toDate string
	cmd := &cobra.Command{
		Use:   "teams",
		Short: "Team analytics (POST /api/v2/account/analytics/team-analytics/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := analyticsBody(cmd, fromDate, toDate)
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, "team analytics", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiV2AccountAnalyticsTeamAnalyticsWithBodyWithResponse(ctx, jsonContentType, bodyReader(body))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&fromDate, "from", "", "from_date YYYY-MM-DD (convenience flag)")
	cmd.Flags().StringVar(&toDate, "to", "", "to_date YYYY-MM-DD (convenience flag)")
	return cmd
}

func newAnalyticsUsersCmd() *cobra.Command {
	var fromDate, toDate string
	cmd := &cobra.Command{
		Use:   "users",
		Short: "User analytics (POST /api/v2/account/analytics/user-analytics/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := analyticsBody(cmd, fromDate, toDate)
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, "user analytics", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiV2AccountAnalyticsUserAnalyticsWithBodyWithResponse(ctx, jsonContentType, bodyReader(body))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&fromDate, "from", "", "from_date YYYY-MM-DD (convenience flag)")
	cmd.Flags().StringVar(&toDate, "to", "", "to_date YYYY-MM-DD (convenience flag)")
	return cmd
}
