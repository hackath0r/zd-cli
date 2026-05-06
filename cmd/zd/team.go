package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/output"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage teams (CRUD + members + permissions + incident roles + task templates)",
	}
	cmd.AddCommand(
		newTeamListCmd(),
		newTeamGetCmd(),
		newTeamCreateCmd(),
		newTeamUpdateCmd(),
		newTeamDeleteCmd(),
		newTeamMemberCmd(),
		newTeamPermissionCmd(),
		newTeamRoleCmd(),
		newTeamTaskTemplateCmd(),
	)
	return cmd
}

func teamTableSpec() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"unique_id", "name", "owner", "created"},
		Row: func(item any) []string {
			t, ok := item.(zenduty.Team)
			if !ok {
				return []string{fmt.Sprintf("%v", item), "", "", ""}
			}
			return []string{
				derefStr(t.UniqueId),
				derefStr(t.Name),
				derefStr(t.Owner),
				derefTimeRFC3339(t.CreationDate),
			}
		},
	}
}

func newTeamListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List teams (GET /api/account/teams/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*[]zenduty.Team, error) {
				resp, err := cli.GetApiAccountTeamsWithResponse(ctx)
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				out, err := decodeList[zenduty.Team](resp.Body)
				if err != nil {
					return nil, zerrors.Networkf(err, "decode teams response")
				}
				return &out, nil
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, *res, teamTableSpec())
		},
	}
}

func newTeamGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id>",
		Short: "Show a team (GET /api/account/teams/{team_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*zenduty.Team, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return resp.JSON200, nil
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, []zenduty.Team{deref(res)}, teamTableSpec())
		},
	}
}

func newTeamCreateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a team (POST /api/account/teams/)",
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
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*zenduty.Team, error) {
				resp, err := cli.PostApiAccountTeamsWithBodyWithResponse(ctx, jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				if resp.JSON201 != nil {
					return resp.JSON201, nil
				}
				return decodeOne[zenduty.Team](resp.Body)
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, []zenduty.Team{deref(res)}, teamTableSpec())
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&name, "name", "", "team name (convenience flag; merged into --body)")
	return cmd
}

func newTeamUpdateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "update <team-id>",
		Short: "Update a team (PUT /api/account/teams/{team_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{"name": nonEmpty(name)})
			if err != nil {
				return err
			}
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*zenduty.Team, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return decodeOne[zenduty.Team](resp.Body)
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, []zenduty.Team{deref(res)}, teamTableSpec())
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&name, "name", "", "team name (convenience flag; merged into --body)")
	return cmd
}

func newTeamDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id>",
		Short: "Delete a team (DELETE /api/account/teams/{team_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*struct{}, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return nil, nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted team %s\n", args[0])
			return nil
		},
	}
}

// ---- members ----------------------------------------------------------------

func newTeamMemberCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "member", Short: "Manage team members"}
	cmd.AddCommand(
		newTeamMemberListCmd(),
		newTeamMemberGetCmd(),
		newTeamMemberAddCmd(),
		newTeamMemberUpdateCmd(),
		newTeamMemberRemoveCmd(),
	)
	return cmd
}

func teamMemberTableSpec() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"unique_id", "user", "role"},
		Row: func(item any) []string {
			b, _ := json.Marshal(item)
			var m map[string]any
			_ = json.Unmarshal(b, &m)
			user, role := "", ""
			if v, ok := m["user"].(string); ok {
				user = v
			}
			if v, ok := m["role"].(float64); ok {
				role = fmt.Sprintf("%d", int(v))
			} else if v, ok := m["role"].(string); ok {
				role = v
			}
			id, _ := m["unique_id"].(string)
			return []string{id, user, role}
		},
	}
}

func newTeamMemberListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List team members (GET /api/account/teams/{team_id}/members/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*[]map[string]any, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdMembersWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				out, err := decodeList[map[string]any](resp.Body)
				if err != nil {
					return nil, err
				}
				return &out, nil
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, *res, teamMemberTableSpec())
		},
	}
}

func newTeamMemberGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <member-id>",
		Short: "Show a team member (GET /api/account/teams/{team_id}/members/{member_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*map[string]any, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdMembersMemberIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return decodeOne[map[string]any](resp.Body)
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, []map[string]any{deref(res)}, teamMemberTableSpec())
		},
	}
}

func newTeamMemberAddCmd() *cobra.Command {
	var user string
	var role int
	cmd := &cobra.Command{
		Use:   "add <team-id>",
		Short: "Add a team member (POST /api/account/teams/{team_id}/members/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"user": nonEmpty(user),
				"role": ifNonZero(role),
			})
			if err != nil {
				return err
			}
			_, _, err = callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*struct{}, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdMembersWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, err
				}
				return nil, requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse))
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added member to team %s\n", args[0])
			return nil
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&user, "user", "", "user username (convenience flag)")
	cmd.Flags().IntVar(&role, "role", 0, "team role integer (convenience flag)")
	return cmd
}

func newTeamMemberUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <member-id>",
		Short: "Update a team member (PUT /api/account/teams/{team_id}/members/{member_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			_, _, err = callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*struct{}, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdMembersMemberIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
				if err != nil {
					return nil, err
				}
				return nil, requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse))
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated member %s on team %s\n", args[1], args[0])
			return nil
		},
	}
	addBodyFlags(cmd)
	return cmd
}

func newTeamMemberRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <team-id> <member-id>",
		Short: "Remove a team member (DELETE /api/account/teams/{team_id}/members/{member_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*struct{}, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdMembersMemberIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, err
				}
				return nil, requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse))
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed member %s from team %s\n", args[1], args[0])
			return nil
		},
	}
}

// ---- permissions ------------------------------------------------------------

func newTeamPermissionCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "permission", Short: "Manage team permissions"}
	cmd.AddCommand(newTeamPermissionGetCmd(), newTeamPermissionUpdateCmd())
	return cmd
}

func newTeamPermissionGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id>",
		Short: "Show team permissions (GET /api/account/teams/{team_id}/permissions/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*zenduty.TeamPermission, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdPermissionsWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return resp.JSON200, nil
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, deref(res), nil)
		},
	}
}

func newTeamPermissionUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id>",
		Short: "Update team permissions (PUT /api/account/teams/{team_id}/permissions/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required (JSON object with account_permissions array)")
			}
			_, _, err = callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, _ zenduty.Config) (*struct{}, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdPermissionsWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
				if err != nil {
					return nil, err
				}
				return nil, requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse))
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated permissions on team %s\n", args[0])
			return nil
		},
	}
	addBodyFlags(cmd)
	return cmd
}

// ---- incident roles ---------------------------------------------------------

func newTeamRoleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "role", Short: "Manage team incident roles"}
	cmd.AddCommand(
		newTeamRoleListCmd(),
		newTeamRoleGetCmd(),
		newTeamRoleCreateCmd(),
		newTeamRoleUpdateCmd(),
		newTeamRoleDeleteCmd(),
	)
	return cmd
}

func newTeamRoleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List incident roles (GET /api/account/teams/{team_id}/roles/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdRolesResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdRolesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTeamRoleGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <role-id>",
		Short: "Show an incident role (GET /api/account/teams/{team_id}/roles/{role_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdRolesIncidentRoleIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdRolesIncidentRoleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTeamRoleCreateCmd() *cobra.Command {
	var title, description string
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create an incident role (POST /api/account/teams/{team_id}/roles/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"title":       nonEmpty(title),
				"description": nonEmpty(description),
			})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, fmt.Sprintf("created incident role on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdRolesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&title, "title", "", "role title (convenience flag)")
	cmd.Flags().StringVar(&description, "description", "", "role description (convenience flag)")
	return cmd
}

func newTeamRoleUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <role-id>",
		Short: "Update an incident role (PUT /api/account/teams/{team_id}/roles/{role_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated incident role %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdRolesIncidentRoleIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newTeamRoleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <role-id>",
		Short: "Delete an incident role (DELETE /api/account/teams/{team_id}/roles/{role_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted incident role %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdRolesIncidentRoleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- task templates ---------------------------------------------------------

func newTeamTaskTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "task-template", Short: "Manage team task templates"}
	cmd.AddCommand(
		newTeamTaskTemplateListCmd(),
		newTeamTaskTemplateGetCmd(),
		newTeamTaskTemplateCreateCmd(),
		newTeamTaskTemplateUpdateCmd(),
		newTeamTaskTemplateDeleteCmd(),
	)
	return cmd
}

func newTeamTaskTemplateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List task templates (GET /api/account/teams/{team_id}/task_templates/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdTaskTemplatesResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdTaskTemplatesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTeamTaskTemplateGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <template-id>",
		Short: "Show a task template (GET /api/account/teams/{team_id}/task_templates/{template_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdTaskTemplatesTaskTemplateIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdTaskTemplatesTaskTemplateIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newTeamTaskTemplateCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create a task template (POST /api/account/teams/{team_id}/task_templates/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created task template on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdTaskTemplatesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newTeamTaskTemplateUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <template-id>",
		Short: "Update a task template (PUT /api/account/teams/{team_id}/task_templates/{template_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated task template %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdTaskTemplatesTaskTemplateIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newTeamTaskTemplateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <template-id>",
		Short: "Delete a task template (DELETE /api/account/teams/{team_id}/task_templates/{template_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted task template %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdTaskTemplatesTaskTemplateIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}
