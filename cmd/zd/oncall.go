package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hackath0r/zd-cli/internal/config"
	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/output"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

func newOncallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "oncall",
		Aliases: []string{"on-call"},
		Short:   "Look up who's currently on call",
	}
	cmd.AddCommand(
		newOncallListCmd(),
		newOncallWhoCmd(),
		newOncallNowCmd(),
	)
	return cmd
}

// onCallRow is the flat representation we render. We keep it free of any
// generated types so json/yaml/table output is consistent across the v1
// and v2 oncall endpoints.
type onCallRow struct {
	Team             string `json:"team" yaml:"team"`
	EscalationPolicy string `json:"escalation_policy" yaml:"escalation_policy"`
	RuleDelayMinutes int    `json:"rule_delay_minutes" yaml:"rule_delay_minutes"`
	User             string `json:"user" yaml:"user"`
	UserEmail        string `json:"user_email" yaml:"user_email"`
}

func oncallTable() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"team", "escalation_policy", "rule_delay_minutes", "user", "email"},
		Row: func(item any) []string {
			r := item.(onCallRow)
			return []string{r.Team, r.EscalationPolicy, fmt.Sprintf("%d", r.RuleDelayMinutes), r.User, r.UserEmail}
		},
	}
}

func newOncallListCmd() *cobra.Command {
	var team string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the oncall roster for a team (GET /api/account/teams/{team_id}/oncall)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			team, err := resolveTeam(cmd, team)
			if err != nil {
				return err
			}
			rows, err := fetchOncallV1(cmd, team)
			if err != nil {
				return err
			}
			return renderResult(cmd, rows, oncallTable())
		},
	}
	cmd.Flags().StringVar(&team, "team", "", "team unique_id (default: profile.default_team)")
	return cmd
}

func newOncallWhoCmd() *cobra.Command {
	var team string
	cmd := &cobra.Command{
		Use:   "who",
		Short: "Print the current on-call user(s) for a team",
		RunE: func(cmd *cobra.Command, _ []string) error {
			team, err := resolveTeam(cmd, team)
			if err != nil {
				return err
			}
			rows, err := fetchOncallV1(cmd, team)
			if err != nil {
				return err
			}
			// "Who" returns only rule 0 (immediate / primary) entries.
			var primary []onCallRow
			for _, r := range rows {
				if r.RuleDelayMinutes == 0 {
					primary = append(primary, r)
				}
			}
			if len(primary) == 0 {
				primary = rows
			}
			return renderResult(cmd, primary, oncallTable())
		},
	}
	cmd.Flags().StringVar(&team, "team", "", "team unique_id (default: profile.default_team)")
	return cmd
}

func newOncallNowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "now",
		Short: "Alias for `oncall who` using the default team from the profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			team, err := resolveTeam(cmd, "")
			if err != nil {
				return err
			}
			rows, err := fetchOncallV1(cmd, team)
			if err != nil {
				return err
			}
			var primary []onCallRow
			for _, r := range rows {
				if r.RuleDelayMinutes == 0 {
					primary = append(primary, r)
				}
			}
			if len(primary) == 0 {
				primary = rows
			}
			return renderResult(cmd, primary, oncallTable())
		},
	}
	return cmd
}

// resolveTeam picks a team unique_id with this precedence:
//  1. explicit --team flag value
//  2. profile.DefaultTeam in the active config profile
func resolveTeam(cmd *cobra.Command, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	v := viperFromCmd(cmd)
	cfgPath := v.GetString(flagConfig)
	file, err := config.Load(cfgPath)
	if err == nil {
		profile := v.GetString(flagProfile)
		if p, _, err := file.Resolve(profile); err == nil && p.DefaultTeam != "" {
			return p.DefaultTeam, nil
		}
	}
	return "", zerrors.Usagef("--team is required (no default_team in profile)")
}

func fetchOncallV1(cmd *cobra.Command, team string) ([]onCallRow, error) {
	res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*[]zenduty.EscalationPolicyOnCallV1, error) {
		resp, err := cli.GetApiAccountTeamsTeamIdOncallWithResponse(ctx, team)
		if err != nil {
			return nil, err
		}
		if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
			return nil, err
		}
		out, err := decodeList[zenduty.EscalationPolicyOnCallV1](resp.Body)
		if err != nil {
			return nil, err
		}
		return &out, nil
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	rows := make([]onCallRow, 0, len(*res))
	for _, item := range *res {
		row := onCallRow{}
		if item.Team != nil {
			row.Team = strings.TrimSpace(derefStr(item.Team.Name))
		}
		if item.EscalationPolicy != nil {
			row.EscalationPolicy = derefStr(item.EscalationPolicy.Name)
		}
		if item.User != nil {
			row.User = strings.TrimSpace(fmt.Sprintf("%s %s", derefStr(item.User.FirstName), derefStr(item.User.LastName)))
			if row.User == "" {
				row.User = derefStr(item.User.Username)
			}
			row.UserEmail = derefStr(item.User.Email)
		}
		rows = append(rows, row)
	}
	return rows, nil
}
