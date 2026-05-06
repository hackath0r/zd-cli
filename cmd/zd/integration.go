package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// `zd integration` wraps the per-service integrations and their child
// transformers (the upstream API name for what some users call
// "alert rules" — see also `zd alert-rule` which is a transparent
// alias preserved for discoverability).

func newIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: "Manage service integrations and transformers",
	}
	cmd.AddCommand(
		newIntegrationListCmd(),
		newIntegrationGetCmd(),
		newIntegrationCreateCmd(),
		newIntegrationUpdateCmd(),
		newIntegrationDeleteCmd(),
		newTransformerCmd(),
	)
	return cmd
}

func newIntegrationListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id> <service-id>",
		Short: "List integrations on a service (GET /api/account/teams/{team_id}/services/{service_id}/integrations/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newIntegrationGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <service-id> <integration-id>",
		Short: "Show an integration (GET /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdWithResponse(ctx, args[0], args[1], args[2])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newIntegrationCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id> <service-id>",
		Short: "Create an integration (POST /api/account/teams/{team_id}/services/{service_id}/integrations/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created integration on service %s", args[1]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdServicesServiceIdIntegrationsWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newIntegrationUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <service-id> <integration-id>",
		Short: "Update an integration (PUT /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated integration %s", args[2]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdWithBodyWithResponse(ctx, args[0], args[1], args[2], jsonContentType, bodyReader(body))
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

func newIntegrationDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <service-id> <integration-id>",
		Short: "Delete an integration (DELETE /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted integration %s", args[2]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdWithResponse(ctx, args[0], args[1], args[2])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- transformers (a.k.a. alert rules) -------------------------------------

func newTransformerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transformer",
		Aliases: []string{"alert-rule", "rule"},
		Short:   "Manage integration transformers / alert rules",
	}
	cmd.AddCommand(
		newTransformerListCmd(),
		newTransformerGetCmd(),
		newTransformerCreateCmd(),
		newTransformerUpdateCmd(),
		newTransformerDeleteCmd(),
	)
	return cmd
}

func newTransformerListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id> <service-id> <integration-id>",
		Short: "List transformers (GET .../integrations/{integration_id}/transformers/)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdTransformersResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdTransformersWithResponse(ctx, args[0], args[1], args[2])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTransformerGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <service-id> <integration-id> <transformer-id>",
		Short: "Show a transformer (GET .../transformers/{transformer_id}/)",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdTransformersTransformersIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdTransformersTransformersIdWithResponse(ctx, args[0], args[1], args[2], args[3])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTransformerCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id> <service-id> <integration-id>",
		Short: "Create a transformer (POST .../integrations/{integration_id}/transformers/)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created transformer on integration %s", args[2]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdTransformersWithBodyWithResponse(ctx, args[0], args[1], args[2], jsonContentType, bodyReader(body))
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

func newTransformerUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <service-id> <integration-id> <transformer-id>",
		Short: "Update a transformer (PATCH .../transformers/{transformer_id}/)",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated transformer %s", args[3]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PatchApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdTransformersTransformersIdWithBodyWithResponse(ctx, args[0], args[1], args[2], args[3], jsonContentType, bodyReader(body))
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

func newTransformerDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <service-id> <integration-id> <transformer-id>",
		Short: "Delete a transformer (DELETE .../transformers/{transformer_id}/)",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted transformer %s", args[3]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdServicesServiceIdIntegrationsIntegrationIdTransformersTransformersIdWithResponse(ctx, args[0], args[1], args[2], args[3])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// `zd alert-rule` is a top-level alias for `zd integration transformer`
// to preserve discoverability for users who think of these as alert
// rules. We register it directly so users don't have to know about
// the `integration transformer` path.

func newAlertRuleCmd() *cobra.Command {
	cmd := newTransformerCmd()
	cmd.Use = "alert-rule"
	cmd.Aliases = []string{"alert-rules", "rule"}
	cmd.Short = "Alias for `zd integration transformer` (alert / event rules on a service integration)"
	return cmd
}
