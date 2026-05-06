package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// `zd user` is sugar over `zd account member`. The Zenduty API doesn't
// have a separate /users/ collection — list/get/update are aliases for
// the members endpoints, but contacts / forwarding-rules /
// notification-rules / customroles do live under /api/account/users/{username}/.

func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users (members, contacts, forwarding rules, notification rules, custom roles)",
	}
	cmd.AddCommand(
		newUserListCmd(),
		newUserGetCmd(),
		newUserUpdateCmd(),
		newUserContactCmd(),
		newUserForwardingRuleCmd(),
		newUserNotificationRuleCmd(),
		newUserCustomRoleCmd(),
	)
	return cmd
}

// `user list/get/update` proxy to the same /api/account/members endpoints.

func newUserListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List users (GET /api/account/members/)",
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

func newUserGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <username>",
		Short: "Show a user (GET /api/account/members/{username}/)",
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

func newUserUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <username>",
		Short: "Update a user (PUT /api/account/members/{username}/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body or --data is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated user %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
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

// ---- user contacts ----------------------------------------------------------

func newUserContactCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "contact", Short: "Manage user contact methods"}
	cmd.AddCommand(
		newUserContactListCmd(),
		newUserContactGetCmd(),
		newUserContactCreateCmd(),
		newUserContactDeleteCmd(),
	)
	return cmd
}

func newUserContactListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <username>",
		Short: "List user contact methods (GET /api/account/users/{username}/contacts/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountUsersUsernameContactsResponse, error) {
				resp, err := cli.GetApiAccountUsersUsernameContactsWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newUserContactGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <username> <contact-id>",
		Short: "Show a user contact method (GET /api/account/users/{username}/contacts/{contact_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountUsersUsernameContactsContactIdResponse, error) {
				resp, err := cli.GetApiAccountUsersUsernameContactsContactIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newUserContactCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <username>",
		Short: "Add a user contact method (POST /api/account/users/{username}/contacts/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("added contact for %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountUsersUsernameContactsWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newUserContactDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <username> <contact-id>",
		Short: "Delete a user contact method (DELETE /api/account/users/{username}/contacts/{contact_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted contact %s for user %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountUsersUsernameContactsContactIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- forwarding rules -------------------------------------------------------

func newUserForwardingRuleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "forwarding-rule", Short: "Manage user forwarding rules"}
	cmd.AddCommand(
		newUserForwardingListCmd(),
		newUserForwardingGetCmd(),
		newUserForwardingCreateCmd(),
		newUserForwardingDeleteCmd(),
	)
	return cmd
}

func newUserForwardingListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <username>",
		Short: "List user forwarding rules (GET /api/account/users/{username}/forwarding_rules/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountUsersUsernameForwardingRulesResponse, error) {
				resp, err := cli.GetApiAccountUsersUsernameForwardingRulesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newUserForwardingGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <username> <rule-id>",
		Short: "Show a user forwarding rule (GET /api/account/users/{username}/forwarding_rules/{rule_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountUsersUsernameForwardingRulesForwardingRuleIdResponse, error) {
				resp, err := cli.GetApiAccountUsersUsernameForwardingRulesForwardingRuleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newUserForwardingCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <username>",
		Short: "Add a forwarding rule (POST /api/account/users/{username}/forwarding_rules/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("added forwarding rule for %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountUsersUsernameForwardingRulesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newUserForwardingDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <username> <rule-id>",
		Short: "Delete a forwarding rule (DELETE /api/account/users/{username}/forwarding_rules/{rule_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted forwarding rule %s for user %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountUsersUsernameForwardingRulesForwardingRuleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- notification rules -----------------------------------------------------

func newUserNotificationRuleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "notification-rule", Short: "Manage user notification rules"}
	cmd.AddCommand(
		newUserNotifListCmd(),
		newUserNotifGetCmd(),
		newUserNotifCreateCmd(),
		newUserNotifDeleteCmd(),
	)
	return cmd
}

func newUserNotifListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <username>",
		Short: "List notification rules (GET /api/account/users/{username}/notification_rules/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountUsersUsernameNotificationRulesResponse, error) {
				resp, err := cli.GetApiAccountUsersUsernameNotificationRulesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newUserNotifGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <username> <rule-id>",
		Short: "Show a notification rule (GET /api/account/users/{username}/notification_rules/{rule_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountUsersUsernameNotificationRulesNotificationRuleIdResponse, error) {
				resp, err := cli.GetApiAccountUsersUsernameNotificationRulesNotificationRuleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newUserNotifCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <username>",
		Short: "Add a notification rule (POST /api/account/users/{username}/notification_rules/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("added notification rule for %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountUsersUsernameNotificationRulesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newUserNotifDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <username> <rule-id>",
		Short: "Delete a notification rule (DELETE /api/account/users/{username}/notification_rules/{rule_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted notification rule %s for user %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountUsersUsernameNotificationRulesNotificationRuleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- user custom roles ------------------------------------------------------

func newUserCustomRoleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "custom-role", Short: "Manage user-level custom-role assignments"}
	cmd.AddCommand(newUserCustomRoleListCmd(), newUserCustomRoleAddCmd())
	return cmd
}

func newUserCustomRoleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <username>",
		Short: "List user custom roles (GET /api/account/users/{username}/customroles/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountUsersUsernameCustomrolesResponse, error) {
				resp, err := cli.GetApiAccountUsersUsernameCustomrolesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newUserCustomRoleAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <username>",
		Short: "Assign a custom role to a user (POST /api/account/users/{username}/customroles/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required (e.g. {\"custom_role\": \"<role-id>\"})")
			}
			return runWriteRaw(cmd, fmt.Sprintf("assigned custom role to user %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountUsersUsernameCustomrolesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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
