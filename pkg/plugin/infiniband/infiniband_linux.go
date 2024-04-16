// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package infiniband contains the Retina infiniband plugin. It gathers infiniband statistics and debug status parameters.
package infiniband

import (
	"context"
	"sync"
	"time"

	hubblev1 "github.com/cilium/cilium/pkg/hubble/api/v1"
	kcfg "github.com/microsoft/retina/pkg/config"
	"github.com/microsoft/retina/pkg/log"
	"github.com/microsoft/retina/pkg/plugin/api"
	"go.uber.org/zap"
)

// New creates a infiniband plugin.
func New(cfg *kcfg.Config) api.Plugin {
	return &infiniband{
		cfg: cfg,
		l:   log.Logger().Named(string(Name)),
	}
}

func (lu *infiniband) Name() string {
	return string(Name)
}

func (lu *infiniband) Generate(ctx context.Context) error { //nolint //implementing iface
	return nil
}

func (lu *infiniband) Compile(ctx context.Context) error { //nolint // implementing iface
	return nil
}

func (lu *infiniband) Init() error {
	lu.l.Info("Initializing infiniband plugin...")
	return nil
}

func (lu *infiniband) Start(ctx context.Context) error {
	lu.isRunning = true
	return lu.run(ctx)
}

func (lu *infiniband) SetupChannel(ch chan *hubblev1.Event) error { // nolint // impl. iface
	lu.l.Warn("Plugin does not support SetupChannel", zap.String("plugin", string(Name)))
	return nil
}

func (lu *infiniband) run(ctx context.Context) error {
	lu.l.Info("Running infiniband plugin...")
	ticker := time.NewTicker(lu.cfg.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			lu.l.Info("Context is done, infiniband will stop running")
			return nil
		case <-ticker.C:
			var wg sync.WaitGroup

			infinibandReader := NewInfinibandReader()
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := infinibandReader.readAndUpdate()
				if err != nil {
					lu.l.Error("Reading infiniband stats failed", zap.Error(err))
				}
			}()

			wg.Wait()
		}
	}
}

func (lu *infiniband) Stop() error {
	if !lu.isRunning {
		return nil
	}
	lu.l.Info("Stopping infiniband plugin...")
	lu.isRunning = false
	return nil
}
