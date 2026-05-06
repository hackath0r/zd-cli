package main

import (
	"github.com/spf13/cobra"
)

// This file wires placeholder cobra commands for every Zenduty / Xurrent
// IMR resource that is in the OpenAPI spec but not yet wrapped by zd.
//
// Why bother? Discoverability. Users running `zd --help` or
// `zd <resource> --help` see the full surface, so they know what is
// coming and can either:
//
//  1. Open a "feature: API coverage" issue (see .github/ISSUE_TEMPLATE),
//  2. Drop down to `curl` for the operation they need today, or
//  3. Pin to a specific zd minor version when the wrapper lands.
//
// Each leaf node uses notImplemented() from util.go which returns a
// uniform exit-2 (usage) error message pointing at the project's
// issue tracker. The exact API path is included so it shows up in
// help / --output json error payloads.
//
// As CRUDs are implemented in follow-up releases, the leaf will move
// from `RunE: notImplemented(...)` to a real RunE. The user-facing
// command name and flags stay stable so scripts written against the
// stub keep working.

// newStubCmd returns a placeholder cobra command with a uniform
// description suffix and a not-implemented RunE.
func newStubCmd(use, short, api string) *cobra.Command {
	return &cobra.Command{
		Use:                use,
		Short:              short,
		Long:               short + "\n\nNot yet wrapped by zd; the upstream API is " + api + ". Track progress and request prioritization at https://github.com/hackath0r/zd-cli/issues.",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: false,
		SilenceUsage:       true,
		RunE:               notImplemented(api),
	}
}

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "account", Short: "Manage account-wide resources"}
	cmd.AddCommand(
		newStubCmd("members", "List or invite account members", "GET/POST /api/account/members/"),
		newStubCmd("roles", "Manage account roles", "GET/POST /api/account/roles/"),
		newStubCmd("custom-roles", "Manage custom roles", "GET/POST /api/account/custom-roles/"),
		newStubCmd("invitations", "List or send account invitations", "GET/POST /api/account/invitations/"),
	)
	return cmd
}

func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "user", Short: "Manage users (members of the current account)"}
	cmd.AddCommand(
		newStubCmd("list", "List users", "GET /api/account/members/"),
		newStubCmd("get", "Show a single user", "GET /api/account/members/{username}/"),
		newStubCmd("update", "Update a user", "PUT /api/account/members/{username}/"),
	)
	return cmd
}

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "team", Short: "Manage teams"}
	cmd.AddCommand(
		newStubCmd("list", "List teams", "GET /api/account/teams/"),
		newStubCmd("get", "Show a team", "GET /api/account/teams/{team_id}/"),
		newStubCmd("create", "Create a team", "POST /api/account/teams/"),
		newStubCmd("update", "Update a team", "PUT /api/account/teams/{team_id}/"),
		newStubCmd("delete", "Delete a team", "DELETE /api/account/teams/{team_id}/"),
	)
	return cmd
}

func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "schedule", Short: "Manage on-call schedules"}
	cmd.AddCommand(
		newStubCmd("list", "List schedules", "GET /api/account/teams/{team_id}/schedules/"),
		newStubCmd("get", "Show a schedule", "GET /api/account/teams/{team_id}/schedules/{schedule_id}/"),
		newStubCmd("create", "Create a schedule", "POST /api/account/teams/{team_id}/schedules/"),
		newStubCmd("update", "Update a schedule", "PUT /api/account/teams/{team_id}/schedules/{schedule_id}/"),
		newStubCmd("delete", "Delete a schedule", "DELETE /api/account/teams/{team_id}/schedules/{schedule_id}/"),
		newStubCmd("override", "Add a schedule override", "POST /api/v2/account/teams/{team_id}/schedules/{schedule_id}/overrides/"),
	)
	return cmd
}

func newServiceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "service", Short: "Manage services"}
	cmd.AddCommand(
		newStubCmd("list", "List services", "GET /api/account/teams/{team_id}/services/"),
		newStubCmd("get", "Show a service", "GET /api/account/teams/{team_id}/services/{service_id}/"),
		newStubCmd("create", "Create a service", "POST /api/account/teams/{team_id}/services/"),
		newStubCmd("update", "Update a service", "PUT /api/account/teams/{team_id}/services/{service_id}/"),
		newStubCmd("delete", "Delete a service", "DELETE /api/account/teams/{team_id}/services/{service_id}/"),
	)
	return cmd
}

func newEscalationPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "escalation-policy",
		Aliases: []string{"ep"},
		Short:   "Manage escalation policies",
	}
	cmd.AddCommand(
		newStubCmd("list", "List escalation policies", "GET /api/account/teams/{team_id}/escalation_policies/"),
		newStubCmd("get", "Show an escalation policy", "GET /api/account/teams/{team_id}/escalation_policies/{ep_id}/"),
		newStubCmd("create", "Create an escalation policy", "POST /api/account/teams/{team_id}/escalation_policies/"),
		newStubCmd("update", "Update an escalation policy", "PUT /api/account/teams/{team_id}/escalation_policies/{ep_id}/"),
		newStubCmd("delete", "Delete an escalation policy", "DELETE /api/account/teams/{team_id}/escalation_policies/{ep_id}/"),
	)
	return cmd
}

func newPriorityCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "priority", Short: "Manage incident priorities"}
	cmd.AddCommand(
		newStubCmd("list", "List priorities", "GET /api/account/priorities/"),
		newStubCmd("create", "Create a priority", "POST /api/account/priorities/"),
		newStubCmd("update", "Update a priority", "PUT /api/account/priorities/{priority_id}/"),
		newStubCmd("delete", "Delete a priority", "DELETE /api/account/priorities/{priority_id}/"),
	)
	return cmd
}

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage account-level tags (incident tag definitions)",
	}
	cmd.AddCommand(
		newStubCmd("list", "List tags", "GET /api/account/tags/"),
		newStubCmd("create", "Create a tag", "POST /api/account/tags/"),
		newStubCmd("delete", "Delete a tag", "DELETE /api/account/tags/{tag_id}/"),
	)
	return cmd
}

func newSLOCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "slo", Short: "Manage service level objectives"}
	cmd.AddCommand(
		newStubCmd("list", "List SLOs", "GET /api/account/teams/{team_id}/slo/"),
		newStubCmd("create", "Create an SLO", "POST /api/account/teams/{team_id}/slo/"),
		newStubCmd("update", "Update an SLO", "PUT /api/account/teams/{team_id}/slo/{slo_id}/"),
		newStubCmd("delete", "Delete an SLO", "DELETE /api/account/teams/{team_id}/slo/{slo_id}/"),
	)
	return cmd
}

func newPostmortemCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "postmortem", Short: "Manage postmortems"}
	cmd.AddCommand(
		newStubCmd("list", "List postmortems", "GET /api/account/teams/{team_id}/postmortems/"),
		newStubCmd("get", "Show a postmortem", "GET /api/account/teams/{team_id}/postmortems/{postmortem_id}/"),
		newStubCmd("create", "Create a postmortem from incident(s)", "POST /api/account/teams/{team_id}/postmortems/"),
		newStubCmd("update", "Update a postmortem", "PUT /api/account/teams/{team_id}/postmortems/{postmortem_id}/"),
		newStubCmd("delete", "Delete a postmortem", "DELETE /api/account/teams/{team_id}/postmortems/{postmortem_id}/"),
	)
	return cmd
}

func newMaintenanceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "maintenance", Short: "Schedule and manage maintenance windows"}
	cmd.AddCommand(
		newStubCmd("list", "List maintenance windows", "GET /api/account/teams/{team_id}/maintenance/"),
		newStubCmd("create", "Schedule a maintenance window", "POST /api/account/teams/{team_id}/maintenance/"),
		newStubCmd("update", "Update a maintenance window", "PUT /api/account/teams/{team_id}/maintenance/{maintenance_id}/"),
		newStubCmd("delete", "Cancel a maintenance window", "DELETE /api/account/teams/{team_id}/maintenance/{maintenance_id}/"),
	)
	return cmd
}

func newRouterCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "router", Short: "Manage event routers (global and team)"}
	cmd.AddCommand(
		newStubCmd("list", "List routers", "GET /api/router/"),
		newStubCmd("create", "Create a router", "POST /api/router/"),
		newStubCmd("update", "Update a router", "PUT /api/router/{router_id}/"),
		newStubCmd("delete", "Delete a router", "DELETE /api/router/{router_id}/"),
	)
	return cmd
}

func newAlertRuleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "alert-rule", Short: "Manage alert (event) rules on services"}
	cmd.AddCommand(
		newStubCmd("list", "List alert rules", "GET /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/alert_rules/"),
		newStubCmd("create", "Create an alert rule", "POST /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/alert_rules/"),
		newStubCmd("update", "Update an alert rule", "PUT /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/alert_rules/{rule_id}/"),
		newStubCmd("delete", "Delete an alert rule", "DELETE /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/alert_rules/{rule_id}/"),
	)
	return cmd
}

func newIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "integration", Short: "Manage service integrations"}
	cmd.AddCommand(
		newStubCmd("list", "List integrations on a service", "GET /api/account/teams/{team_id}/services/{service_id}/integrations/"),
		newStubCmd("create", "Create an integration", "POST /api/account/teams/{team_id}/services/{service_id}/integrations/"),
		newStubCmd("update", "Update an integration", "PUT /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/"),
		newStubCmd("delete", "Delete an integration", "DELETE /api/account/teams/{team_id}/services/{service_id}/integrations/{integration_id}/"),
	)
	return cmd
}

func newAnalyticsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "analytics", Short: "Read incident analytics"}
	cmd.AddCommand(
		newStubCmd("incidents", "Incident analytics", "POST /api/account/incidents/analytics/"),
		newStubCmd("services", "Service analytics", "POST /api/account/services/analytics/"),
		newStubCmd("teams", "Team analytics", "POST /api/account/teams/analytics/"),
		newStubCmd("oncall", "On-call analytics", "GET /api/account/oncall/analytics/"),
	)
	return cmd
}
