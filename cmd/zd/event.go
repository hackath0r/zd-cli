package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/output"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// newEventCmd builds the `zd event` command tree.
//
// Events flow through the modern generic ingestion endpoint:
//
//	POST /integration/{account_id}/generic/{integration_key}/
//
// which returns a trace_id (no incident yet); the alert is processed
// asynchronously and may or may not create an incident based on the
// integration's alert_rules. `zd event status <trace>` polls the trace
// to see what happened. `zd event fire --wait` chains the two.
func newEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "event",
		Aliases: []string{"events"},
		Short:   "Send and trace events through Zenduty integrations",
	}
	cmd.AddCommand(
		newEventFireCmd(),
		newEventAckCmd(),
		newEventResolveCmd(),
		newEventStatusCmd(),
	)
	return cmd
}

type eventFlags struct {
	integrationKey string
	alertType      string
	message        string
	summary        string
	entityID       string
	urls           []string
	payloadJSON    string
	priority       string
	sla            string
	escalation     string
	tags           []string
	wait           bool
	waitTimeout    time.Duration
}

func registerEventFlags(cmd *cobra.Command, f *eventFlags) {
	cmd.Flags().StringVar(&f.integrationKey, "integration-key", "", "integration key (UUID; required)")
	cmd.Flags().StringVar(&f.alertType, "alert-type", "critical", "critical | acknowledged | resolved | error | warning | info")
	cmd.Flags().StringVarP(&f.message, "message", "m", "", "alert message (required)")
	cmd.Flags().StringVarP(&f.summary, "summary", "s", "", "alert summary")
	cmd.Flags().StringVar(&f.entityID, "entity-id", "", "deduplication key (idempotent retries; recommended)")
	cmd.Flags().StringSliceVar(&f.urls, "url", nil, "related URL (repeatable)")
	cmd.Flags().StringVar(&f.payloadJSON, "payload", "", "extra structured payload as JSON")
	cmd.Flags().StringVar(&f.priority, "priority", "", "team priority unique_id (optional)")
	cmd.Flags().StringVar(&f.sla, "sla", "", "team SLA unique_id (optional)")
	cmd.Flags().StringVar(&f.escalation, "escalation-policy", "", "escalation policy unique_id (optional)")
	cmd.Flags().StringSliceVar(&f.tags, "tag", nil, "tag unique_id (repeatable)")
	cmd.Flags().BoolVar(&f.wait, "wait", false, "after firing, poll status until terminal")
	cmd.Flags().DurationVar(&f.waitTimeout, "wait-timeout", 60*time.Second, "max time to wait for terminal status")
}

func eventStatusTable() *output.TableSpec {
	return &output.TableSpec{
		Headers: []string{"trace_id", "status", "alert", "incident", "incident_created"},
		Row: func(item any) []string {
			s := item.(eventStatusRow)
			return []string{
				s.TraceID,
				s.Status,
				s.Alert,
				s.IncidentNumber,
				fmt.Sprintf("%v", s.IncidentCreated),
			}
		},
	}
}

type eventStatusRow struct {
	TraceID         string `json:"trace_id" yaml:"trace_id"`
	Status          string `json:"status" yaml:"status"`
	Alert           string `json:"alert" yaml:"alert"`
	IncidentNumber  string `json:"incident_number" yaml:"incident_number"`
	IncidentCreated bool   `json:"incident_created" yaml:"incident_created"`
}

func newEventFireCmd() *cobra.Command {
	f := &eventFlags{}
	cmd := &cobra.Command{
		Use:   "fire",
		Short: "Send an event to an integration (POST /integration/{account_id}/generic/{key}/)",
		Long: `Send an event through Zenduty's modern ingestion endpoint. The endpoint
returns a trace_id immediately; the alert is then processed
asynchronously and may or may not create an incident based on the
integration's alert rules. Use --wait to block until the trace reaches
a terminal status (completed or failed).

Idempotency: pass --entity-id to deduplicate retries on the Zenduty
side; subsequent events with the same entity_id will reference the same
alert (and incident, if one was created).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEventFire(cmd, f)
		},
	}
	registerEventFlags(cmd, f)
	return cmd
}

func newEventAckCmd() *cobra.Command {
	f := &eventFlags{alertType: "acknowledged"}
	cmd := &cobra.Command{
		Use:   "ack",
		Short: "Send an acknowledged event (alert_type=acknowledged)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			f.alertType = "acknowledged"
			return runEventFire(cmd, f)
		},
	}
	registerEventFlags(cmd, f)
	return cmd
}

func newEventResolveCmd() *cobra.Command {
	f := &eventFlags{alertType: "resolved"}
	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Send a resolved event (alert_type=resolved)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			f.alertType = "resolved"
			return runEventFire(cmd, f)
		},
	}
	registerEventFlags(cmd, f)
	return cmd
}

func runEventFire(cmd *cobra.Command, f *eventFlags) error {
	if f.integrationKey == "" {
		return zerrors.Usagef("--integration-key is required")
	}
	if f.message == "" {
		return zerrors.Usagef("--message is required")
	}
	if !validAlertType(f.alertType) {
		return zerrors.Usagef("--alert-type=%q invalid; want one of critical|acknowledged|resolved|error|warning|info", f.alertType)
	}

	body := zenduty.GenericEventPayload{
		AlertType:        f.alertType,
		Message:          f.message,
		Summary:          strPtrOrNil(f.summary),
		EntityId:         strPtrOrNil(f.entityID),
		Priority:         strPtrOrNil(f.priority),
		Sla:              strPtrOrNil(f.sla),
		EscalationPolicy: strPtrOrNil(f.escalation),
	}
	if len(f.urls) > 0 {
		body.Urls = &f.urls
	}
	if len(f.tags) > 0 {
		body.Tags = &f.tags
	}
	if f.payloadJSON != "" {
		var pl map[string]any
		if err := json.Unmarshal([]byte(f.payloadJSON), &pl); err != nil {
			return zerrors.Usagef("--payload is not valid JSON: %v", err)
		}
		body.Payload = &pl
	}

	res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.EventV3, error) {
		accountID, err := cfg.AccountIDOrError()
		if err != nil {
			return nil, &zerrors.ExitError{Code: zerrors.CodeConfig, Msg: err.Error()}
		}
		resp, err := cli.PostIntegrationAccountIdGenericIntegrationKeyWithResponse(ctx, accountID, f.integrationKey, body)
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
	traceID := derefStr(res.TraceId)

	row := eventStatusRow{TraceID: traceID, Status: derefStr(res.Message)}
	if !f.wait {
		return renderResult(cmd, []eventStatusRow{row}, eventStatusTable())
	}

	final, err := waitForTraceStatus(cmd, traceID, f.waitTimeout)
	if err != nil {
		return err
	}
	return renderResult(cmd, []eventStatusRow{final}, eventStatusTable())
}

func validAlertType(s string) bool {
	switch strings.ToLower(s) {
	case "critical", "acknowledged", "resolved", "error", "warning", "info":
		return true
	default:
		return false
	}
}

func newEventStatusCmd() *cobra.Command {
	var trace string
	var watch bool
	var watchTimeout time.Duration
	cmd := &cobra.Command{
		Use:   "status <trace-id>",
		Short: "Get the status of a fired event (GET /api/alert/status/{trace_id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			trace = args[0]
			if watch {
				row, err := waitForTraceStatus(cmd, trace, watchTimeout)
				if err != nil {
					return err
				}
				return renderResult(cmd, []eventStatusRow{row}, eventStatusTable())
			}
			row, err := fetchTraceStatus(cmd, trace)
			if err != nil {
				return err
			}
			return renderResult(cmd, []eventStatusRow{row}, eventStatusTable())
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "poll until terminal status")
	cmd.Flags().DurationVar(&watchTimeout, "watch-timeout", 60*time.Second, "max time to wait")
	return cmd
}

func fetchTraceStatus(cmd *cobra.Command, traceID string) (eventStatusRow, error) {
	res, _, err := callAPI(cmd, func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*zenduty.AlertStatus, error) {
		resp, err := cli.GetApiAlertStatusTraceIdWithResponse(ctx, traceID)
		if err != nil {
			return nil, err
		}
		if err := requireOK(resp, resp.Body, urlFromResp(resp.HTTPResponse)); err != nil {
			return nil, err
		}
		return resp.JSON200, nil
	})
	if err != nil {
		return eventStatusRow{}, err
	}
	return alertStatusToRow(traceID, res), nil
}

func alertStatusToRow(traceID string, s *zenduty.AlertStatus) eventStatusRow {
	row := eventStatusRow{TraceID: traceID}
	if s == nil {
		return row
	}
	row.Status = derefStr(s.Status)
	row.Alert = derefStr(s.Alert)
	if s.Incident != nil && s.Incident.IncidentNumber != nil {
		row.IncidentNumber = fmt.Sprintf("%d", *s.Incident.IncidentNumber)
	}
	row.IncidentCreated = derefBool(s.IsIncidentCreated)
	return row
}

func waitForTraceStatus(cmd *cobra.Command, traceID string, timeout time.Duration) (eventStatusRow, error) {
	deadline := time.Now().Add(timeout)
	delay := 500 * time.Millisecond
	maxDelay := 5 * time.Second

	for {
		row, err := fetchTraceStatus(cmd, traceID)
		if err != nil {
			return row, err
		}
		switch strings.ToLower(row.Status) {
		case "completed", "failed":
			return row, nil
		}
		if time.Now().After(deadline) {
			return row, &zerrors.ExitError{
				Code: zerrors.CodeAPI,
				Msg:  fmt.Sprintf("timed out waiting for trace %s to reach terminal state (last status %q)", traceID, row.Status),
			}
		}
		select {
		case <-cmd.Context().Done():
			return row, cmd.Context().Err()
		case <-time.After(delay):
		}
		if delay < maxDelay {
			delay *= 2
		}
	}
}
