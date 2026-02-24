package opsgenie

import (
	"context"
	"log/slog"
	"time"

	"github.com/giantswarm/oka/pkg/config"
)

// Service is a service for fetching alerts from OpsGenie.
type Service struct {
	alertClient *AlertClient
	query       string
	interval    time.Duration
}

// NewService creates a new OpsGenie service.
func NewService(conf config.OpsGenie) (*Service, error) {
	alertClient, err := NewAlertClient(conf.APIUrl, conf.EnvVar)
	if err != nil {
		return nil, err
	}

	query, err := TemplateQuery(conf.QueryString, conf.Team)
	if err != nil {
		return nil, err
	}

	s := &Service{
		alertClient: alertClient,
		interval:    conf.Interval,
		query:       query,
	}

	return s, nil
}

// Start starts the OpsGenie service, which periodically fetches alerts and
// sends them to the provided channel.
func (s *Service) Start(ctx context.Context, queryChan chan<- any) {
	slog.Info("OpsGenie service started", "interval", s.interval, "query", s.query)
	defer slog.Info("OpsGenie service stopped")

	ticker := time.Tick(s.interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			slog.Info("Fetching alerts from OpsGenie")

			alerts, err := s.alertClient.ListAlerts(ctx, s.query)
			if err != nil {
				slog.Error("Failed to fetch alerts from OpsGenie", "error", err)
				continue
			}

			if len(alerts) == 0 {
				slog.Info("No new alerts found in OpsGenie")
				continue
			}

			count := 0
			for _, alert := range alerts {
				if alert.Acknowledged {
					continue // Skip acknowledged alerts
				}

				a, err := s.alertClient.GetAlert(ctx, alert.Id)
				if err != nil {
					slog.Warn("Failed to get alert from OpsGenie", "id", alert.Id, "error", err)
					continue
				}

				// TODO: Maybe find the installation related to this alert and pass an
				// installation parameter in order to restrict the available kube contexts.

				// Send the alert to the channel for further processing.
				queryChan <- a
				count++
			}

			slog.Info("Fetched new alerts from OpsGenie", "new", count, "total", len(alerts))
		}
	}
}
