package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage account-wide resources (members, custom roles, invites, integration metadata)",
	}
	cmd.AddCommand(
		newAccountMemberCmd(),
		newAccountCustomRoleCmd(),
		newAccountInviteCmd(),
		newAccountDeleteUserCmd(),
		newAccountRegenerateIntegrationKeyCmd(),
		newAccountIntegrationMetadataCmd(),
	)
	return cmd
}

// ---- members ----------------------------------------------------------------

func newAccountMemberCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "member", Short: "Manage account members"}
	cmd.AddCommand(
		newAccountMemberListCmd(),
		newAccountMemberGetCmd(),
		newAccountMemberUpdateCmd(),
	)
	return cmd
}

func newAccountMemberListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List account members (GET /api/account/members/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountMembersResponse, error) {
				resp, err := cli.GetApiAccountMembersWithResponse(ctx)
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newAccountMemberGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <username>",
		Short: "Show an account member (GET /api/account/members/{username}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountMembersUsernameResponse, error) {
				resp, err := cli.GetApiAccountMembersUsernameWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newAccountMemberUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <username>",
		Short: "Update an account member (PUT /api/account/members/{username}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated member %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountMembersUsernameWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

// ---- custom roles -----------------------------------------------------------

func newAccountCustomRoleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "custom-role", Short: "Manage account-level custom roles"}
	cmd.AddCommand(
		newAccountCustomRoleListCmd(),
		newAccountCustomRoleGetCmd(),
		newAccountCustomRoleCreateCmd(),
		newAccountCustomRoleUpdateCmd(),
		newAccountCustomRoleDeleteCmd(),
	)
	return cmd
}

func newAccountCustomRoleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List custom roles (GET /api/account/customroles/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountCustomrolesResponse, error) {
				resp, err := cli.GetApiAccountCustomrolesWithResponse(ctx)
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newAccountCustomRoleGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <role-id>",
		Short: "Show a custom role (GET /api/account/customroles/{role_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountCustomrolesCustomRoleIdResponse, error) {
				resp, err := cli.GetApiAccountCustomrolesCustomRoleIdWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newAccountCustomRoleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom role (POST /api/account/customroles/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, "created custom role", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountCustomrolesWithBodyWithResponse(ctx, jsonContentType, bodyReader(body))
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

func newAccountCustomRoleUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <role-id>",
		Short: "Update a custom role (PUT /api/account/customroles/{role_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated custom role %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountCustomrolesCustomRoleIdWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newAccountCustomRoleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <role-id>",
		Short: "Delete a custom role (DELETE /api/account/customroles/{role_id}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted custom role %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountCustomrolesCustomRoleIdWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- invite -----------------------------------------------------------------

func newAccountInviteCmd() *cobra.Command {
	var email string
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Send an invitation (POST /api/account/api_invite/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"email": nonEmpty(email),
			})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, "sent invitation", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountApiInviteWithBodyWithResponse(ctx, jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&email, "email", "", "invitee email (convenience flag)")
	return cmd
}

// ---- delete user ------------------------------------------------------------

func newAccountDeleteUserCmd() *cobra.Command {
	var username string
	cmd := &cobra.Command{
		Use:   "delete-user",
		Short: "Delete a user (POST /api/account/deleteuser/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"username": nonEmpty(username),
			})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, "user deleted", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountDeleteuserWithBodyWithResponse(ctx, jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&username, "username", "", "username to delete (convenience flag)")
	return cmd
}

// ---- regenerate-integration-key --------------------------------------------

func newAccountRegenerateIntegrationKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate-integration-key",
		Short: "Regenerate an integration key (POST /api/account/regenerate_integration_key/)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required (e.g. {\"integration_key\": \"...\"})")
			}
			return runWriteRaw(cmd, "integration key regenerated", func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountRegenerateIntegrationKeyWithBodyWithResponse(ctx, jsonContentType, bodyReader(body))
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

// ---- integration metadata ---------------------------------------------------

func newAccountIntegrationMetadataCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "integration-metadata <integration-key>",
		Short: "Show integration metadata (GET /api/account/integration_metadata/{integration_key}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountIntegrationMetadataIntegrationKeyResponse, error) {
				resp, err := cli.GetApiAccountIntegrationMetadataIntegrationKeyWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}
