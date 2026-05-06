package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

func newServiceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "service", Short: "Manage services on a team"}
	cmd.AddCommand(
		newServiceListCmd(),
		newServiceGetCmd(),
		newServiceCreateCmd(),
		newServiceUpdateCmd(),
		newServiceDeleteCmd(),
	)
	return cmd
}

func newServiceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List services (GET /api/account/teams/{team_id}/services/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdServicesResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdServicesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newServiceGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <service-id>",
		Short: "Show a service (GET /api/account/teams/{team_id}/services/{service_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdServicesServiceIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdServicesServiceIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newServiceCreateCmd() *cobra.Command {
	var name, description, escalationPolicy string
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create a service (POST /api/account/teams/{team_id}/services/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"name":              nonEmpty(name),
				"description":       nonEmpty(description),
				"escalation_policy": nonEmpty(escalationPolicy),
			})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, fmt.Sprintf("created service on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdServicesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&name, "name", "", "service name (convenience flag)")
	cmd.Flags().StringVar(&description, "description", "", "service description (convenience flag)")
	cmd.Flags().StringVar(&escalationPolicy, "escalation-policy", "", "escalation policy unique_id (convenience flag)")
	return cmd
}

func newServiceUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <service-id>",
		Short: "Update a service (PUT /api/account/teams/{team_id}/services/{service_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated service %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdServicesServiceIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	return cmd
}

func newServiceDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <service-id>",
		Short: "Delete a service (DELETE /api/account/teams/{team_id}/services/{service_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted service %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdServicesServiceIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}
