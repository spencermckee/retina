// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package infiniband

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/microsoft/retina/pkg/log"
	"github.com/microsoft/retina/pkg/metrics"
	"go.uber.org/zap"
)

const (
	pathInfiniband            = "/home/spencermckee/retina-ib-test/infiniband" //"/sys/class/infiniband"
	pathDebugStatusParameters = "/home/spencermckee/retina-ib-test/net"        //"sys/class/net"
)

func NewInfinibandReader() *InfinibandReader {
	return &InfinibandReader{
		l:                log.Logger().Named(string("InfinibandReader")),
		counterStats:     make(map[CounterStat]uint64),
		statusParamStats: make(map[StatusParam]uint64),
	}
}

type InfinibandReader struct { // nolint // clearer naming
	l                *log.ZapLogger
	counterStats     map[CounterStat]uint64
	statusParamStats map[StatusParam]uint64
}

func (ir *InfinibandReader) readAndUpdate() error {
	if err := ir.readCounterStats(pathInfiniband); err != nil {
		return err
	}

	if err := ir.readStatusParamStats(pathDebugStatusParameters); err != nil {
		return err
	}

	ir.updateMetrics()
	ir.l.Debug("Done reading and updating stats")

	return nil
}

func (ir *InfinibandReader) readCounterStats(path string) error {
	ir.counterStats = make(map[CounterStat]uint64)
	devices, err := os.ReadDir(path)
	if err != nil {
		ir.l.Error("error reading dir:", zap.Error(err))
		return err // nolint std. fmt.
	}
	for _, device := range devices {
		portsPath := filepath.Join(path, device.Name(), "ports")
		ports, err := os.ReadDir(portsPath)
		if err != nil {
			ir.l.Error("error reading dir:", zap.Error(err))
			continue
		}
		for _, port := range ports {
			countersPath := filepath.Join(portsPath, port.Name(), "counters")
			counters, err := os.ReadDir(countersPath)
			if err != nil {
				ir.l.Error("error reading dir:", zap.Error(err))
				continue
			}
			for _, counter := range counters {
				counterPath := filepath.Join(countersPath, counter.Name())
				val, err := os.ReadFile(counterPath)
				if err != nil {
					ir.l.Error("Error while reading infiniband file: \n", zap.Error(err))
					continue
				}
				num, err := strconv.ParseUint(strings.TrimSpace(string(val)), 10, 64)
				if err != nil {
					ir.l.Error("error parsing string:", zap.Error(err))
					return err // nolint std. fmt.
				}
				ir.counterStats[CounterStat{Name: counter.Name(), Device: device.Name(), Port: port.Name()}] = num
			}

		}
	}
	return nil
}

func (ir *InfinibandReader) readStatusParamStats(path string) error {
	ifaces, err := os.ReadDir(path)
	if err != nil {
		ir.l.Error("error reading dir:", zap.Error(err))
		return err // nolint std. fmt.
	}
	ir.statusParamStats = make(map[StatusParam]uint64)
	for _, iface := range ifaces {
		statusParamsPath := filepath.Join(path, iface.Name(), "debug")
		statusParams, err := os.ReadDir(statusParamsPath)
		if err != nil {
			ir.l.Error("error parsing string:", zap.Error(err))
			continue
		}
		for _, statusParam := range statusParams {
			statusParamPath := filepath.Join(statusParamsPath, statusParam.Name())
			val, err := os.ReadFile(statusParamPath)
			if err != nil {
				ir.l.Error("Error while reading infiniband path file: \n", zap.Error(err))
				continue
			}
			num, err := strconv.ParseUint(string(val), 10, 64)
			if err != nil {
				ir.l.Error("Error while reading infiniband file: \n", zap.Error(err))
				return err // nolint std. fmt.
			}
			ir.statusParamStats[StatusParam{Name: statusParam.Name(), Iface: iface.Name()}] = num

		}
	}
	return nil
}

func (ir *InfinibandReader) updateMetrics() {
	if ir.counterStats == nil {
		ir.l.Info("No stats found")
		return
	}
	if ir.statusParamStats == nil {
		ir.l.Info("No status param stats found")
		return
	}

	// Adding counter stats
	for counter, val := range ir.counterStats {
		metrics.InfinibandCounterStats.WithLabelValues(counter.Name, counter.Device, counter.Port).Set(float64(val))
	}

	// Adding status params
	for statusParam, val := range ir.statusParamStats {
		metrics.InfinibandStatusParams.WithLabelValues(statusParam.Name, statusParam.Iface).Set(float64(val))
	}
}
