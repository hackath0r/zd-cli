package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "schedule", Short: "Manage schedules and overrides on a team"}
	cmd.AddCommand(
		newScheduleListCmd(),
		newScheduleGetCmd(),
		newScheduleCreateCmd(),
		newScheduleUpdateCmd(),
		newScheduleDeleteCmd(),
		newScheduleOverrideCmd(),
	)
	return cmd
}

func newScheduleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id>",
		Short: "List schedules (GET /api/account/teams/{team_id}/schedules/)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdSchedulesResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdSchedulesWithResponse(ctx, args[0])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newScheduleGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <team-id> <schedule-id>",
		Short: "Show a schedule (GET /api/account/teams/{team_id}/schedules/{schedule_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiAccountTeamsTeamIdSchedulesScheduleIdResponse, error) {
				resp, err := cli.GetApiAccountTeamsTeamIdSchedulesScheduleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newScheduleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <team-id>",
		Short: "Create a schedule (POST /api/account/teams/{team_id}/schedules/)",
		Long: `Create a schedule. Body shape (see Zenduty docs):
  {"name": "Primary Rotation", "time_zone": "Asia/Kolkata", "layers": [...]}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("created schedule on team %s", args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiAccountTeamsTeamIdSchedulesWithBodyWithResponse(ctx, args[0], jsonContentType, bodyReader(body))
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

func newScheduleUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <team-id> <schedule-id>",
		Short: "Update a schedule (PUT /api/account/teams/{team_id}/schedules/{schedule_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return zerrors.Usagef("--body is required")
			}
			return runWriteRaw(cmd, fmt.Sprintf("updated schedule %s on team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PutApiAccountTeamsTeamIdSchedulesScheduleIdWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(body))
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

func newScheduleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <team-id> <schedule-id>",
		Short: "Delete a schedule (DELETE /api/account/teams/{team_id}/schedules/{schedule_id}/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, fmt.Sprintf("deleted schedule %s from team %s", args[1], args[0]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.DeleteApiAccountTeamsTeamIdSchedulesScheduleIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
}

// ---- overrides --------------------------------------------------------------

func newScheduleOverrideCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "override",
		Short: "Manage schedule overrides (one-off shift swaps)",
	}
	cmd.AddCommand(
		newScheduleOverrideListCmd(),
		newScheduleOverrideAddCmd(),
	)
	return cmd
}

func newScheduleOverrideListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <team-id> <schedule-id>",
		Short: "List schedule overrides (GET /api/v2/account/teams/{team_id}/schedules/{schedule_id}/overrides/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRaw(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses) ([]byte, *zenduty.GetApiV2AccountTeamsTeamIdSchedulesScheduleIdOverridesResponse, error) {
				resp, err := cli.GetApiV2AccountTeamsTeamIdSchedulesScheduleIdOverridesWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, nil, err
				}
				return resp.Body, resp, nil
			})
		},
	}
}

func newScheduleOverrideAddCmd() *cobra.Command {
	var user, startTime, endTime string
	cmd := &cobra.Command{
		Use:   "add <team-id> <schedule-id>",
		Short: "Create an override (POST /api/v2/account/teams/{team_id}/schedules/{schedule_id}/overrides/)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONBody(cmd, "")
			if err != nil {
				return err
			}
			merged, err := mergeBody(body, map[string]any{
				"user":       nonEmpty(user),
				"start_time": nonEmpty(startTime),
				"end_time":   nonEmpty(endTime),
			})
			if err != nil {
				return err
			}
			return runWriteRaw(cmd, fmt.Sprintf("added override on schedule %s", args[1]), func(ctx context.Context, cli *zenduty.ClientWithResponses) (httpResponse, []byte, error) {
				resp, err := cli.PostApiV2AccountTeamsTeamIdSchedulesScheduleIdOverridesWithBodyWithResponse(ctx, args[0], args[1], jsonContentType, bodyReader(merged))
				if err != nil {
					return nil, nil, err
				}
				return resp, resp.Body, nil
			})
		},
	}
	addBodyFlags(cmd)
	cmd.Flags().StringVar(&user, "user", "", "user username (convenience flag)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "start_time (ISO8601, convenience flag)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "end_time (ISO8601, convenience flag)")
	return cmd
}
