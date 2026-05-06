package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// `zd router` wraps the v2 event router and rulesets API.

func newRouterCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "router", Short: "Manage event routers and rulesets (v2)"}
	cmd.AddCommand(
		newRouterListCmd(),
		newRouterGetCmd(),
		newRouterCreateCmd(),
		newRouterUpdateCmd(),
		newRouterDeleteCmd(),
		newRulesetCmd(),
	)
	return cmd
}

func newRouterListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List event routers (GET /api/v2/account/events/router/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiV2AccountEventsRouterResponse, error) {
				resp, err := cli.GetApiV2AccountEventsRouterWithResponse(ctx)
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newRouterGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <router-id>",
		Short: "Show an event router (GET /api/v2/account/events/router/{router_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiV2AccountEventsRouterRouterIdResponse, error) {
				resp, err := cli.GetApiV2AccountEventsRouterRouterIdWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newRouterCreateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an event router (POST /api/v2/account/events/router/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{"name": nonEmpty(name)})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, "created event router", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiV2AccountEventsRouterWithBodyWithResponse(ctx, jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&name, "name", "", "router name (convenience flag)")
	return cmd
}

func newRouterUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <router-id>",
		Short: "Update an event router (PUT /api/v2/account/events/router/{router_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated event router %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiV2AccountEventsRouterRouterIdWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newRouterDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <router-id>",
		Short: "Delete an event router (DELETE /api/v2/account/events/router/{router_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted event router %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiV2AccountEventsRouterRouterIdWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- rulesets ---------------------------------------------------------------

func newRulesetCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "ruleset", Short: "Manage rulesets on an event router"}
	cmd.AddCommand(
		newRulesetListCmd(),
		newRulesetGetCmd(),
		newRulesetCreateCmd(),
		newRulesetReorderCmd(),
		newRulesetUpdateCmd(),
		newRulesetDeleteCmd(),
	)
	return cmd
}

func newRulesetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <router-id>",
		Short: "List rulesets (GET /api/v2/account/events/router/{router_id}/rulesets/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiV2AccountEventsRouterRouterIdRulesetsResponse, error) {
				resp, err := cli.GetApiV2AccountEventsRouterRouterIdRulesetsWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newRulesetGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <router-id> <ruleset-id>",
		Short: "Show a ruleset (GET /api/v2/account/events/router/{router_id}/rulesets/{ruleset_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiV2AccountEventsRouterRouterIdRulesetsRulesetIdResponse, error) {
				resp, err := cli.GetApiV2AccountEventsRouterRouterIdRulesetsRulesetIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newRulesetCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <router-id>",
		Short: "Create a ruleset (POST /api/v2/account/events/router/{router_id}/rulesets/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created ruleset on router %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiV2AccountEventsRouterRouterIdRulesetsWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newRulesetReorderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reorder <router-id>",
		Short: "Reorder rulesets (PUT /api/v2/account/events/router/{router_id}/rulesets/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required (array of {unique_id, position})")
			}
			return runWriteRaw(cmd, fmt.Sprintf("reordered rulesets on router %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiV2AccountEventsRouterRouterIdRulesetsWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newRulesetUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <router-id> <ruleset-id>",
		Short: "Update a ruleset (PUT /api/v2/account/events/router/{router_id}/rulesets/{ruleset_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated ruleset %s on router %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiV2AccountEventsRouterRouterIdRulesetsRulesetIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newRulesetDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <router-id> <ruleset-id>",
		Short: "Delete a ruleset (DELETE /api/v2/account/events/router/{router_id}/rulesets/{ruleset_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted ruleset %s from router %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiV2AccountEventsRouterRouterIdRulesetsRulesetIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}
