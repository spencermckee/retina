// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package infiniband

import (
	"testing"

	"github.com/microsoft/retina/pkg/log"
	"github.com/microsoft/retina/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

var (
	MockGaugeVec   *metrics.MockIGaugeVec
	MockCounterVec *metrics.MockICounterVec
)

func TestNewInfinibandReader(t *testing.T) {
	log.SetupZapLogger(log.GetDefaultLogOpts())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nr := NewInfinibandReader()
	assert.NotNil(t, nr)
}

func InitalizeMetricsForTesting(ctrl *gomock.Controller) {
	metricsLogger := log.Logger().Named("metrics")
	metricsLogger.Info("Initializing metrics for testing")

	MockGaugeVec = metrics.NewMockIGaugeVec(ctrl)
	metrics.InfinibandCounterStats = MockGaugeVec //nolint:typecheck // no type check
	metrics.InfinibandStatusParams = MockGaugeVec
}

//nolint:testifylint // not making linter changes to preserve exact behavior
func TestReadCounterStats(t *testing.T) {
	log.SetupZapLogger(log.GetDefaultLogOpts())
	tests := []struct {
		name     string
		filePath string
		result   *CounterStat
		wantErr  bool
	}{
		{
			name:     "test correct",
			filePath: "testdata/infiniband",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			nr := NewInfinibandReader()
			InitalizeMetricsForTesting(ctrl)

			testmetric := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "testmetric",
				Help: "testmetric",
			})

			MockGaugeVec.EXPECT().WithLabelValues(gomock.Any()).Return(testmetric).AnyTimes()

			assert.NotNil(t, nr)
			err := nr.readCounterStats(tt.filePath)
			if tt.wantErr {
				assert.NotNil(t, err, "Expected error but got nil")
			} else {
				assert.Nil(t, err, "Expected nil but got err")
				assert.NotNil(t, nr.counterStats, "Expected data got nil")
				for _, val := range nr.counterStats {
					assert.Equal(t, val, uint64(1))
				}
				assert.Equal(t, len(nr.counterStats), 6, "Read values are not equal to expected")
				nr.updateMetrics()
			}
		})
	}
}

func TestReadStatusParamStats(t *testing.T) {
	log.SetupZapLogger(log.GetDefaultLogOpts())
	tests := []struct {
		name     string
		filePath string
		result   *StatusParam
		wantErr  bool
	}{
		{
			name:     "test correct",
			filePath: "testdata/net",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			nr := NewInfinibandReader()
			InitalizeMetricsForTesting(ctrl)

			testmetric := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "testmetric",
				Help: "testmetric",
			})

			MockGaugeVec.EXPECT().WithLabelValues(gomock.Any()).Return(testmetric).AnyTimes()

			assert.NotNil(t, nr)
			err := nr.readStatusParamStats(tt.filePath)
			if tt.wantErr {
				assert.NotNil(t, err, "Expected error but got nil") // nolint std. fmt.
			} else {
				assert.Nil(t, err, "Expected nil but got err") // nolint std. fmt.
				assert.NotNil(t, nr.statusParamStats, "Expected data got nil")
				for _, val := range nr.statusParamStats {
					assert.Equal(t, uint64(1), val)
				}
				assert.Equal(t, len(nr.statusParamStats), 4, "Read values are not equal to expected") // nolint // no issue

				nr.updateMetrics()
			}
		})
	}
}
