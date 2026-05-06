package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// This file wraps every team-scoped resource that follows the same
// CRUD shape: GET list, GET one, POST create, PUT update, DELETE.
// Each resource has its own dedicated path under /api/account/teams/{team_id}/
// and we delegate to dedicated newXxxCmd() functions below. Generic-ish
// factories aren't possible here because the cobra/oapi-codegen API
// surface uses concrete typed methods (different functions per resource),
// so we hand-write five RunE bodies per resource and let the shared
// runListRaw / runGetRaw / runWriteRaw / runDelete helpers in util.go
// keep the boilerplate minimal.

// ---- escalation-policy ------------------------------------------------------

func newEscalationPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "escalation-policy",
		Aliases: []string{"ep"},
		Short:   "Manage escalation policies",
	}
	cmd.AddCommand(
		newEPListCmd(),
		newEPGetCmd(),
		newEPCreateCmd(),
		newEPUpdateCmd(),
		newEPDeleteCmd(),
	)
	return cmd
}

func newEPListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List escalation policies (GET /api/account/teams/{team_id}/escalation_policies/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdEscalationPoliciesResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdEscalationPoliciesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newEPGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <ep-id>",
		Short: "Show an escalation policy (GET /api/account/teams/{team_id}/escalation_policies/{ep_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdEscalationPoliciesEscalationPolicyIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdEscalationPoliciesEscalationPolicyIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newEPCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create an escalation policy (POST /api/account/teams/{team_id}/escalation_policies/)",
		Long: `Create an escalation policy. The Zenduty schema is non-trivial (rules + targets);
pass --body @ep.json with the full payload. See the Zenduty docs for the EP schema.

Example body:
  {"name": "Primary", "rules": [{"delay": 0, "targets": [{"type": "user", "user": "alice"}]}]}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created escalation policy on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdEscalationPoliciesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newEPUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <ep-id>",
		Short: "Update an escalation policy (PUT /api/account/teams/{team_id}/escalation_policies/{ep_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated escalation policy %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdEscalationPoliciesEscalationPolicyIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newEPDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <ep-id>",
		Short: "Delete an escalation policy (DELETE /api/account/teams/{team_id}/escalation_policies/{ep_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted escalation policy %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdEscalationPoliciesEscalationPolicyIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- priority ---------------------------------------------------------------

func newPriorityCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "priority", Short: "Manage incident priorities"}
	cmd.AddCommand(
		newPriorityListCmd(),
		newPriorityGetCmd(),
		newPriorityCreateCmd(),
		newPriorityUpdateCmd(),
		newPriorityDeleteCmd(),
	)
	return cmd
}

func newPriorityListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List priorities (GET /api/account/teams/{team_id}/priority/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdPriorityResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdPriorityWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newPriorityGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <priority-id>",
		Short: "Show a priority (GET /api/account/teams/{team_id}/priority/{priority_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdPriorityPriorityIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdPriorityPriorityIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newPriorityCreateCmd() *cobra.Command {
	var name, description, color string
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create a priority (POST /api/account/teams/{team_id}/priority/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"name":        nonEmpty(name),
				"description": nonEmpty(description),
				"color":       nonEmpty(color),
			})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, fmt.Sprintf("created priority on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdPriorityWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&name, "name", "", "priority name (convenience flag)")
	cmd.Flags().StringVar(&description, "description", "", "priority description (convenience flag)")
	cmd.Flags().StringVar(&color, "color", "", "priority color hex (convenience flag)")
	return cmd
}

func newPriorityUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <priority-id>",
		Short: "Update a priority (PUT /api/account/teams/{team_id}/priority/{priority_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated priority %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdPriorityPriorityIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newPriorityDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <priority-id>",
		Short: "Delete a priority (DELETE /api/account/teams/{team_id}/priority/{priority_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted priority %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdPriorityPriorityIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- tags (per-team) --------------------------------------------------------

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags",
		Aliases: []string{"tag"},
		Short:   "Manage incident tag definitions on a team",
	}
	cmd.AddCommand(
		newTagListCmd(),
		newTagGetCmd(),
		newTagCreateCmd(),
		newTagUpdateCmd(),
		newTagDeleteCmd(),
	)
	return cmd
}

func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List tags (GET /api/account/teams/{team_id}/tags/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdTagsResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdTagsWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTagGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <tag-id>",
		Short: "Show a tag (GET /api/account/teams/{team_id}/tags/{tag_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdTagsTagIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdTagsTagIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTagCreateCmd() *cobra.Command {
	var name, color string
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create a tag (POST /api/account/teams/{team_id}/tags/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"name":  nonEmpty(name),
				"color": nonEmpty(color),
			})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, fmt.Sprintf("created tag on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdTagsWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&name, "name", "", "tag name (convenience flag)")
	cmd.Flags().StringVar(&color, "color", "", "tag color (convenience flag)")
	return cmd
}

func newTagUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <tag-id>",
		Short: "Update a tag (PUT /api/account/teams/{team_id}/tags/{tag_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated tag %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdTagsTagIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newTagDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <tag-id>",
		Short: "Delete a tag (DELETE /api/account/teams/{team_id}/tags/{tag_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted tag %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdTagsTagIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- sla --------------------------------------------------------------------

func newSLACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sla",
		Aliases: []string{"slo"}, // historical: 'slo' alias kept for users following older docs
		Short:   "Manage SLAs (Service Level Agreements) on a team",
	}
	cmd.AddCommand(
		newSLAListCmd(),
		newSLAGetCmd(),
		newSLACreateCmd(),
		newSLAUpdateCmd(),
		newSLADeleteCmd(),
	)
	return cmd
}

func newSLAListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List SLAs (GET /api/account/teams/{team_id}/sla/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdSlaResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdSlaWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newSLAGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <sla-id>",
		Short: "Show an SLA (GET /api/account/teams/{team_id}/sla/{sla_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdSlaSlaIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdSlaSlaIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newSLACreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create an SLA (POST /api/account/teams/{team_id}/sla/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created SLA on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdSlaWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newSLAUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <sla-id>",
		Short: "Update an SLA (PUT /api/account/teams/{team_id}/sla/{sla_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated SLA %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdSlaSlaIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newSLADeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <sla-id>",
		Short: "Delete an SLA (DELETE /api/account/teams/{team_id}/sla/{sla_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted SLA %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdSlaSlaIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- postmortem -------------------------------------------------------------

func newPostmortemCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "postmortem", Short: "Manage postmortems on a team"}
	cmd.AddCommand(
		newPostmortemListCmd(),
		newPostmortemGetCmd(),
		newPostmortemCreateCmd(),
		newPostmortemUpdateCmd(),
		newPostmortemDeleteCmd(),
	)
	return cmd
}

func newPostmortemListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List postmortems (GET /api/account/teams/{team_id}/postmortem/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdPostmortemResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdPostmortemWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newPostmortemGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <postmortem-id>",
		Short: "Show a postmortem (GET /api/account/teams/{team_id}/postmortem/{postmortem_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdPostmortemPostmortemIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdPostmortemPostmortemIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newPostmortemCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create a postmortem (POST /api/account/teams/{team_id}/postmortem/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required (Zenduty postmortem schema)")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created postmortem on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdPostmortemWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newPostmortemUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <postmortem-id>",
		Short: "Update a postmortem (PUT /api/account/teams/{team_id}/postmortem/{postmortem_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated postmortem %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdPostmortemPostmortemIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newPostmortemDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <postmortem-id>",
		Short: "Delete a postmortem (DELETE /api/account/teams/{team_id}/postmortem/{postmortem_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted postmortem %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdPostmortemPostmortemIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- maintenance ------------------------------------------------------------

func newMaintenanceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "maintenance", Short: "Manage maintenance windows on a team"}
	cmd.AddCommand(
		newMaintenanceListCmd(),
		newMaintenanceGetCmd(),
		newMaintenanceCreateCmd(),
		newMaintenanceUpdateCmd(),
		newMaintenanceDeleteCmd(),
	)
	return cmd
}

func newMaintenanceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List maintenance windows (GET /api/account/teams/{team_id}/maintenance/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdMaintenanceResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdMaintenanceWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newMaintenanceGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <maintenance-id>",
		Short: "Show a maintenance window (GET /api/account/teams/{team_id}/maintenance/{maintenance_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdMaintenanceMaintenanceIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdMaintenanceMaintenanceIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newMaintenanceCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Schedule a maintenance window (POST /api/account/teams/{team_id}/maintenance/)",
		Long: `Schedule a maintenance window. Body shape (see Zenduty docs for fields):
  {"name": "deploy", "from_time": "...", "to_time": "...", "service_ids": ["..."]}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created maintenance window on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdMaintenanceWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newMaintenanceUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <maintenance-id>",
		Short: "Update a maintenance window (PUT /api/account/teams/{team_id}/maintenance/{maintenance_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated maintenance window %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdMaintenanceMaintenanceIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newMaintenanceDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <maintenance-id>",
		Short: "Cancel a maintenance window (DELETE /api/account/teams/{team_id}/maintenance/{maintenance_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted maintenance window %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdMaintenanceMaintenanceIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}
