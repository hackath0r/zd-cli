package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/output"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// newIncidentCmd builds the `zd incident` command tree. Aliases:
//   - `zd inc` for ergonomic typing
//   - the generated `ximr incident` invocation gets these subcommands
//     for free via the basename detection in newRootCmd.
func newIncidentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "incident",
		Aliases: []string{"inc", "incidents"},
		Short:   "Triage, list, and update incidents",
	}
	cmd.AddCommand(
		newIncidentListCmd(),
		newIncidentGetCmd(),
		newIncidentCreateCmd(),
		newIncidentUpdateCmd(),
		newIncidentAckCmd(),
		newIncidentResolveCmd(),
		newIncidentNoteCmd(),
		newIncidentTagCmd(),
		newIncidentAlertsCmd(),
		newIncidentResponderCmd(),
	)
	return cmd
}

// resolveIncidentUID looks up an incident's unique_id (UUID) given either
// a numeric incident_number or an already-resolved unique_id. PATCH /api/
// incidents/{unique_id} requires unique_id but most user-facing references
// are by incident_number, so this helper is used by ack / resolve / update.
func resolveIncidentUID(ctx context.Context, cli *zenduty.ClientWithResponses, ref string) (string, error) {
	if looksLikeUUID(ref) {
		return ref, nil
	}
	resp, err := cli.GetApiIncidentsIncidentNumberWithResponse(ctx, ref)
	if err != nil {
		return "", zerrors.Networkf(err, "GET /api/incidents/%s", ref)
	}
	if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
		return "", err
	}
	if resp.JSON200 == nil || resp.JSON200.UniqueId == nil || *resp.JSON200.UniqueId == "" {
		return "", zerrors.API(resp.StatusCode(), urlFromResp(resp.HTTPResponse), "no unique_id on incident response")
	}
	return *resp.JSON200.UniqueId, nil
}

func incidentTableSpec() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"number", "status", "urgency", "title", "service", "assigned_to", "created"},
		Row: func(item any) []string {
			switch v := item.(type) {
			case zenduty.IncidentFilterObject:
				return []string{
					fmt.Sprintf("%d", derefInt(v.IncidentNumber)),
					statusName(v.Status),
					"", // IncidentFilterObject schema lacks Urgency
					truncate(derefStr(v.Title), 40),
					serviceName(v.ServiceObject),
					derefStr(v.AssignedToName),
					derefTimeRFC3339(v.CreationDate),
				}
			case zenduty.Incident:
				return []string{
					fmt.Sprintf("%d", derefInt(v.IncidentNumber)),
					statusName(v.Status),
					urgencyName(v.Urgency),
					truncate(derefStr(v.Title), 40),
					serviceName(v.ServiceObject),
					derefStr(v.AssignedToName),
					derefTimeRFC3339(v.CreationDate),
				}
			default:
				return []string{fmt.Sprintf("%v", v), "", "", "", "", "", ""}
			}
		},
	}
}

func serviceName(s *zenduty.ServicePostPut) string {
	if s == nil {
		return ""
	}
	return derefStr(s.Name)
}

func newIncidentListCmd() *cobra.Command {
	var (
		statusFilter     []string
		teamIDs          []string
		serviceIDs       []string
		priorityIDs      []string
		userIDs          []string
		fromDate, toDate string
		page             int
		allTeams         bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incidents (POST /api/incidents/filter)",
		Long: `List incidents using Zenduty's filter endpoint. Defaults to OPEN incidents
(triggered + acknowledged) so the most common use case is one keystroke:

  zd incident list

Status flag values: open (default), all, triggered, acknowledged, resolved.
Pass --status multiple times for multi-status filters.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.IncidentFilter, error) {
				body := zenduty.IncidentFilterPayload{}
				if status, err := normalizeStatusFilter(statusFilter); err != nil {
					return nil, err
				} else if status != nil {
					body.Status = status
				}
				if len(teamIDs) > 0 {
					body.TeamIds = &teamIDs
				}
				if len(serviceIDs) > 0 {
					body.ServiceIds = &serviceIDs
				}
				if len(priorityIDs) > 0 {
					body.PriorityIds = &priorityIDs
				}
				if len(userIDs) > 0 {
					body.UserIds = &userIDs
				}
				if fromDate != "" {
					body.FromDate = &fromDate
				}
				if toDate != "" {
					body.ToDate = &toDate
				}
				if allTeams {
					i := 1
					body.AllTeams = &i
				}
				params := &zenduty.PostApiIncidentsFilterParams{}
				if page > 0 {
					params.Page = &page
				}
				resp, err := cli.PostApiIncidentsFilterWithResponse(ctx, params, body)
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
			if res == nil || res.Results == nil {
				return renderResult(cmd, []zenduty.IncidentFilterObject{}, incidentTableSpec())
			}
			return renderResult(cmd, *res.Results, incidentTableSpec())
		},
	}
	cmd.Flags().StringSliceVar(&statusFilter, "status", []string{"open"}, "status filter: open|all|triggered|acknowledged|resolved (repeatable)")
	cmd.Flags().StringSliceVar(&teamIDs, "team", nil, "filter by team unique_id (repeatable)")
	cmd.Flags().StringSliceVar(&serviceIDs, "service", nil, "filter by service unique_id (repeatable)")
	cmd.Flags().StringSliceVar(&priorityIDs, "priority", nil, "filter by priority unique_id (repeatable)")
	cmd.Flags().StringSliceVar(&userIDs, "user", nil, "filter by assigned user username (repeatable)")
	cmd.Flags().StringVar(&fromDate, "from", "", "from_date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toDate, "to", "", "to_date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&page, "page", 0, "pagination page number")
	cmd.Flags().BoolVar(&allTeams, "all-teams", false, "include incidents across all teams (not just yours)")
	return cmd
}

// normalizeStatusFilter maps friendly --status values into the integer the
// filter API expects. The Zenduty mapping is:
//
//	-1 -> open (triggered + acknowledged)
//	 0 -> all (triggered + acknowledged + resolved)
//	 1 -> triggered, 2 -> acknowledged, 3 -> resolved
//
// We accept comma-separated multi-values from cobra StringSliceVar but
// fold them down to a single integer because the API does not accept a
// list. Multi-status outside of "open" / "all" is rejected with a clear
// error.
func normalizeStatusFilter(values []string) (*int, error) {
	if len(values) == 0 {
		open := -1
		return &open, nil
	}
	if len(values) == 1 {
		switch values[0] {
		case "open", "":
			open := -1
			return &open, nil
		case "all":
			all := 0
			return &all, nil
		case "triggered":
			i := 1
			return &i, nil
		case "acknowledged", "ack":
			i := 2
			return &i, nil
		case "resolved":
			i := 3
			return &i, nil
		default:
			return nil, zerrors.Usagef("--status: unknown value %q (want open|all|triggered|acknowledged|resolved)", values[0])
		}
	}
	return nil, zerrors.Usagef("--status accepts a single value; got %v", values)
}

func newIncidentGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <number>",
		Short: "Show a single incident (GET /api/incidents/{incident_number})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.Incident, error) {
				resp, err := cli.GetApiIncidentsIncidentNumberWithResponse(ctx, args[0])
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
			return renderResult(cmd, []zenduty.Incident{*res}, incidentTableSpec())
		},
	}
}

func newIncidentCreateCmd() *cobra.Command {
	var (
		title, summary, service, escalation, assignedTo, priority, sla string
		urgency                                                        int
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Open a new incident (POST /api/incidents)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" || service == "" {
				return zerrors.Usagef("--title and --service are required")
			}
			body := zenduty.IncidentPayload{
				Title:            title,
				Service:          service,
				Summary:          strPtrOrNil(summary),
				EscalationPolicy: strPtrOrNil(escalation),
				AssignedTo:       strPtrOrNil(assignedTo),
				TeamPriority:     strPtrOrNil(priority),
				Sla:              strPtrOrNil(sla),
			}
			if cmd.Flags().Changed("urgency") {
				body.Status = nil // not used on create
				_ = urgency       // note: IncidentPayload schema exposes no urgency field; flag retained for future use
			}
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.Incident, error) {
				resp, err := cli.PostApiIncidentsWithResponse(ctx, body)
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return resp.JSON201, nil
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, []zenduty.Incident{*res}, incidentTableSpec())
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "incident title (required)")
	cmd.Flags().StringVar(&summary, "summary", "", "incident summary")
	cmd.Flags().StringVar(&service, "service", "", "service unique_id (required)")
	cmd.Flags().StringVar(&escalation, "escalation-policy", "", "escalation policy unique_id")
	cmd.Flags().StringVar(&assignedTo, "assign", "", "assignee username")
	cmd.Flags().StringVar(&priority, "priority", "", "team priority unique_id")
	cmd.Flags().StringVar(&sla, "sla", "", "team SLA unique_id")
	cmd.Flags().IntVar(&urgency, "urgency", 0, "0 (low) or 1 (high)")
	return cmd
}

func newIncidentUpdateCmd() *cobra.Command {
	var (
		title, summary, assignedTo, priority, sla, escalation string
		statusFlag, urgencyFlag                               int
	)
	cmd := &cobra.Command{
		Use:   "update <number>",
		Short: "Update incident fields (PATCH /api/incidents/{unique_id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := zenduty.IncidentUpdatePayload{
				Title:            strPtrOrNil(title),
				Summary:          strPtrOrNil(summary),
				AssignedTo:       strPtrOrNil(assignedTo),
				TeamPriority:     strPtrOrNil(priority),
				Sla:              strPtrOrNil(sla),
				EscalationPolicy: strPtrOrNil(escalation),
			}
			if cmd.Flags().Changed("status") {
				body.Status = &statusFlag
			}
			if cmd.Flags().Changed("urgency") {
				body.Urgency = &urgencyFlag
			}
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.Incident, error) {
				uid, err := resolveIncidentUID(ctx, cli, args[0])
				if err != nil {
					return nil, err
				}
				resp, err := cli.PatchApiIncidentsUniqueIdWithResponse(ctx, uid, body)
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
			return renderResult(cmd, []zenduty.Incident{*res}, incidentTableSpec())
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&summary, "summary", "", "new summary")
	cmd.Flags().StringVar(&assignedTo, "assign", "", "new assignee username")
	cmd.Flags().StringVar(&priority, "priority", "", "team priority unique_id")
	cmd.Flags().StringVar(&sla, "sla", "", "team SLA unique_id")
	cmd.Flags().StringVar(&escalation, "escalation-policy", "", "escalation policy unique_id")
	cmd.Flags().IntVar(&statusFlag, "status", 0, "1=triggered, 2=acknowledged, 3=resolved")
	cmd.Flags().IntVar(&urgencyFlag, "urgency", 0, "0=low, 1=high")
	return cmd
}

func newIncidentAckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ack <number>",
		Short: "Acknowledge an incident (sets status=2)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.Incident, error) {
				uid, err := resolveIncidentUID(ctx, cli, args[0])
				if err != nil {
					return nil, err
				}
				resp, err := cli.PatchApiIncidentsUniqueIdWithResponse(ctx, uid, zenduty.IncidentUpdatePayload{Status: ptrInt(2)})
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
			return renderResult(cmd, []zenduty.Incident{*res}, incidentTableSpec())
		},
	}
}

func newIncidentResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <number>",
		Short: "Resolve an incident (sets status=3)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.Incident, error) {
				uid, err := resolveIncidentUID(ctx, cli, args[0])
				if err != nil {
					return nil, err
				}
				resp, err := cli.PatchApiIncidentsUniqueIdWithResponse(ctx, uid, zenduty.IncidentUpdatePayload{Status: ptrInt(3)})
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
			return renderResult(cmd, []zenduty.Incident{*res}, incidentTableSpec())
		},
	}
}

func newIncidentAlertsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "alerts <number>",
		Short: "List alerts attached to an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*[]zenduty.IncidentAlert, error) {
				resp, err := cli.GetApiIncidentsIncidentNumberAlertsWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				out, err := decodeList[zenduty.IncidentAlert](resp.Body)
				if err != nil {
					return nil, err
				}
				return &out, nil
			})
			if err != nil {
				return err
			}
			list := []zenduty.IncidentAlert{}
			if items != nil {
				list = *items
			}
			return renderResult(cmd, list, &output.TableSpec{
				Headers: []string{"unique_id", "type", "summary", "integration", "created"},
				Row: func(item any) []string {
					a := item.(zenduty.IncidentAlert)
					return []string{
						derefStr(a.UniqueId),
						alertTypeName(a.AlertType),
						truncate(derefStr(a.Summary), 60),
						derefStr(a.Integration),
						derefTimeRFC3339(a.CreationDate),
					}
				},
			})
		},
	}
}

func alertTypeName(t *int) string {
	if t == nil {
		return ""
	}
	switch *t {
	case 0:
		return "info"
	case 1:
		return "warning"
	case 2:
		return "error"
	case 3:
		return "critical"
	case 4:
		return "acknowledged"
	case 5:
		return "resolved"
	default:
		return fmt.Sprintf("%d", *t)
	}
}

// --- notes ----------------------------------------------------------------

func newIncidentNoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "note",
		Aliases: []string{"notes"},
		Short:   "Manage notes on an incident",
	}
	cmd.AddCommand(
		newIncidentNoteListCmd(),
		newIncidentNoteAddCmd(),
		newIncidentNoteUpdateCmd(),
		newIncidentNoteDeleteCmd(),
	)
	return cmd
}

func incidentNoteTable() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"unique_id", "user", "note", "created"},
		Row: func(item any) []string {
			n := item.(zenduty.IncidentNote)
			return []string{
				derefStr(n.UniqueId),
				asString(n.UserName, derefStr(n.User)),
				truncate(derefStr(n.Note), 60),
				derefTimeRFC3339(n.CreationDate),
			}
		},
	}
}

func newIncidentNoteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <number>",
		Short: "List notes on an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*[]zenduty.IncidentNote, error) {
				resp, err := cli.GetApiIncidentsIncidentNumberNoteWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				out, err := decodeList[zenduty.IncidentNote](resp.Body)
				if err != nil {
					return nil, err
				}
				return &out, nil
			})
			if err != nil {
				return err
			}
			list := []zenduty.IncidentNote{}
			if items != nil {
				list = *items
			}
			return renderResult(cmd, list, incidentNoteTable())
		},
	}
}

func newIncidentNoteAddCmd() *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:   "add <number> --message TEXT",
		Short: "Add a note to an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if message == "" {
				return zerrors.Usagef("--message is required")
			}
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.IncidentNote, error) {
				resp, err := cli.PostApiIncidentsIncidentNumberNoteWithResponse(ctx, args[0], zenduty.IncidentNotePayload{Note: &message})
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return resp.JSON201, nil
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, []zenduty.IncidentNote{*res}, incidentNoteTable())
		},
	}
	cmd.Flags().StringVarP(&message, "message", "m", "", "note text (required)")
	return cmd
}

func newIncidentNoteUpdateCmd() *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:   "update <number> <note-id> --message TEXT",
		Short: "Update an existing note",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if message == "" {
				return zerrors.Usagef("--message is required")
			}
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.IncidentNote, error) {
				resp, err := cli.PutApiIncidentsIncidentNumberNoteNoteUniqueIdWithResponse(ctx, args[0], args[1], zenduty.IncidentNotePayload{Note: &message})
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return decodeOne[zenduty.IncidentNote](resp.Body)
			})
			if err != nil {
				return err
			}
			if res == nil {
				return renderResult(cmd, []zenduty.IncidentNote{}, incidentNoteTable())
			}
			return renderResult(cmd, []zenduty.IncidentNote{*res}, incidentNoteTable())
		},
	}
	cmd.Flags().StringVarP(&message, "message", "m", "", "new note text (required)")
	return cmd
}

func newIncidentNoteDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <number> <note-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a note",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*struct{}, error) {
				resp, err := cli.DeleteApiIncidentsIncidentNumberNoteNoteUniqueIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return &struct{}{}, nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted note %s on incident %s\n", args[1], args[0])
			return nil
		},
	}
}

// --- tags -----------------------------------------------------------------

func newIncidentTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tag",
		Aliases: []string{"tags"},
		Short:   "Manage tags on an incident",
	}
	cmd.AddCommand(
		newIncidentTagListCmd(),
		newIncidentTagAddCmd(),
		newIncidentTagDeleteCmd(),
	)
	return cmd
}

func incidentTagTable() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"unique_id", "name", "color", "tag_id", "created"},
		Row: func(item any) []string {
			t := item.(zenduty.IncidentTag)
			return []string{
				derefStr(t.UniqueId),
				derefStr(t.Name),
				derefStr(t.Color),
				derefStr(t.TagId),
				derefTimeRFC3339(t.CreationDate),
			}
		},
	}
}

func newIncidentTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <number>",
		Short: "List tags on an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*[]zenduty.IncidentTag, error) {
				resp, err := cli.GetApiIncidentsIncidentNumberTagsWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				out, err := decodeList[zenduty.IncidentTag](resp.Body)
				if err != nil {
					return nil, err
				}
				return &out, nil
			})
			if err != nil {
				return err
			}
			list := []zenduty.IncidentTag{}
			if items != nil {
				list = *items
			}
			return renderResult(cmd, list, incidentTagTable())
		},
	}
}

func newIncidentTagAddCmd() *cobra.Command {
	var teamTag string
	cmd := &cobra.Command{
		Use:   "add <number> --team-tag <id>",
		Short: "Attach a tag to an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if teamTag == "" {
				return zerrors.Usagef("--team-tag is required (team tag unique_id)")
			}
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.IncidentTag, error) {
				resp, err := cli.PostApiIncidentsIncidentNumberTagsWithResponse(ctx, args[0], zenduty.IncidentTagPayload{TeamTag: &teamTag})
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return resp.JSON201, nil
			})
			if err != nil {
				return err
			}
			return renderResult(cmd, []zenduty.IncidentTag{*res}, incidentTagTable())
		},
	}
	cmd.Flags().StringVar(&teamTag, "team-tag", "", "team tag unique_id (required)")
	return cmd
}

func newIncidentTagDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <number> <tag-unique-id>",
		Aliases: []string{"rm"},
		Short:   "Detach a tag from an incident",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*struct{}, error) {
				resp, err := cli.DeleteApiIncidentsIncidentNumberTagsTagUniqueIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return &struct{}{}, nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed tag %s from incident %s\n", args[1], args[0])
			return nil
		},
	}
}

// --- responders -----------------------------------------------------------

func newIncidentResponderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "responder",
		Aliases: []string{"responders"},
		Short:   "Manage incident responders (v2 API)",
	}
	cmd.AddCommand(
		newIncidentResponderListCmd(),
		newIncidentResponderAddUserCmd(),
		newIncidentResponderAddPolicyCmd(),
		newIncidentResponderRemoveCmd(),
	)
	return cmd
}

func incidentResponderTable() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"unique_id", "kind", "name", "status", "created"},
		Row: func(item any) []string {
			r := item.(zenduty.IncidentResponderDetail)
			kind, name := "user", asString(r.UserName, derefStr(r.User))
			if r.EscalationPolicy != nil {
				kind = "policy"
				name = asString(r.EscalationPolicyName, derefStr(r.EscalationPolicy))
			}
			return []string{
				derefStr(r.UniqueId),
				kind,
				name,
				responderStatusName(r.ResponseStatus),
				derefTimeRFC3339(r.CreationDate),
			}
		},
	}
}

func responderStatusName(s *int) string {
	if s == nil {
		return ""
	}
	switch *s {
	case 1:
		return "accepted"
	case 2:
		return "declined"
	case 3:
		return "pending"
	default:
		return fmt.Sprintf("%d", *s)
	}
}

func newIncidentResponderListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <number>",
		Short: "List responders on an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.IncidentResponderResponse, error) {
				resp, err := cli.GetApiV2IncidentsIncidentNumberRespondersWithResponse(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return decodeOne[zenduty.IncidentResponderResponse](resp.Body)
			})
			if err != nil {
				return err
			}
			items := []zenduty.IncidentResponderDetail{}
			if res != nil && res.Results != nil {
				items = *res.Results
			}
			return renderResult(cmd, items, incidentResponderTable())
		},
	}
}

func newIncidentResponderAddUserCmd() *cobra.Command {
	var user string
	cmd := &cobra.Command{
		Use:   "add-user <number> --user <username>",
		Short: "Page a user as an incident responder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if user == "" {
				return zerrors.Usagef("--user is required")
			}
			var bodyInner zenduty.PostApiV2IncidentsIncidentNumberRespondersJSONBody
			if err := bodyInner.FromIncidentResponderUserPayload(zenduty.IncidentResponderUserPayload{User: user}); err != nil {
				return zerrors.Usagef("encode user payload: %v", err)
			}
			body := zenduty.PostApiV2IncidentsIncidentNumberRespondersJSONRequestBody(bodyInner)
			_, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*struct{}, error) {
				resp, err := cli.PostApiV2IncidentsIncidentNumberRespondersWithResponse(ctx, args[0], body)
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return &struct{}{}, nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "paged user %s on incident %s\n", user, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "user username (required)")
	return cmd
}

func newIncidentResponderAddPolicyCmd() *cobra.Command {
	var policy string
	cmd := &cobra.Command{
		Use:   "add-policy <number> --policy <id>",
		Short: "Page an escalation policy as an incident responder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if policy == "" {
				return zerrors.Usagef("--policy is required")
			}
			var bodyInner zenduty.PostApiV2IncidentsIncidentNumberRespondersJSONBody
			if err := bodyInner.FromIncidentResponderEscalationPolicyPayload(zenduty.IncidentResponderEscalationPolicyPayload{EscalationPolicy: policy}); err != nil {
				return zerrors.Usagef("encode policy payload: %v", err)
			}
			body := zenduty.PostApiV2IncidentsIncidentNumberRespondersJSONRequestBody(bodyInner)
			_, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*struct{}, error) {
				resp, err := cli.PostApiV2IncidentsIncidentNumberRespondersWithResponse(ctx, args[0], body)
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return &struct{}{}, nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "paged escalation policy %s on incident %s\n", policy, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&policy, "policy", "", "escalation policy unique_id (required)")
	return cmd
}

func newIncidentResponderRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <number> <responder-unique-id>",
		Aliases: []string{"rm"},
		Short:   "Remove an incident responder",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*struct{}, error) {
				resp, err := cli.DeleteApiIncidentsIncidentNumberRespondersUniqueIdWithResponse(ctx, args[0], args[1])
				if err != nil {
					return nil, err
				}
				if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
					return nil, err
				}
				return &struct{}{}, nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed responder %s from incident %s\n", args[1], args[0])
			return nil
		},
	}
}
