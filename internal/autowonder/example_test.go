package autowonder_test

import (
	"context"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

// The daemon supplies its own configuration, state directory and model source.
func ExampleEndpoint() {
	connect := func(ctx context.Context, cfg aw.Config, stateDir string, list func(context.Context, string) ([]aw.Model, error)) error {
		client, err := aw.New(cfg)
		if err != nil {
			return err
		}
		pending, err := client.OpenOutbox(stateDir)
		if err != nil {
			return err
		}
		endpoint, err := aw.NewEndpoint(func() aw.Heartbeat {
			return aw.Heartbeat{ProtocolFeatures: []aw.Feature{aw.FeatureQoderModelCatalogV1}}
		}, pending)
		if err != nil {
			return err
		}
		err = aw.On(endpoint, func(ctx context.Context, s *aw.Session, q *aw.QoderModelCatalogRequest) error {
			models, err := list(ctx, q.Provider)
			return endpoint.Send(ctx, s, &aw.QoderModelCatalogResult{
				RequestID: q.RequestID, Provider: q.Provider, Success: err == nil && len(models) > 0, Models: models,
			})
		})
		if err != nil {
			return err
		}
		return client.Run(ctx, 30*time.Second, endpoint)
	}
	_ = connect
}
