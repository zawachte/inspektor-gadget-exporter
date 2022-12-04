package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	commonutils "github.com/inspektor-gadget/inspektor-gadget/cmd/common/utils"
	"github.com/inspektor-gadget/inspektor-gadget/cmd/kubectl-gadget/utils"

	"go.uber.org/zap"

	eventtypes "github.com/inspektor-gadget/inspektor-gadget/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	gadgetv1alpha1 "github.com/inspektor-gadget/inspektor-gadget/pkg/apis/gadget/v1alpha1"

	dnstypes "github.com/inspektor-gadget/inspektor-gadget/pkg/gadgets/trace/dns/types"
	tcptypes "github.com/inspektor-gadget/inspektor-gadget/pkg/gadgets/trace/tcp/types"
)

var (
	dnsResolutions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dns_resolutions_total",
		Help: "Number of DNS Resolutions.",
	}, []string{"node", "qr", "nameserver", "name"})

	tcpResolutions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tcp_resolutions_total",
		Help: "Number of TCP Connections",
	}, []string{"node", "namespace", "pod", "container", "operation", "saddr", "daddr", "dport"})
)

// CollectorOptions
type CollectorOptions struct {
	Logger              *zap.Logger
	KubernetesNamespace string
}

// Collector
type Collector interface {
	Collect(context.Context) error
}

// GadgetCollector
type GadgetCollector struct {
	GadgetName string
	Callback   func(line string, node string)
}

// collector
type collector struct {
	gadgetCollectors    []GadgetCollector
	logger              *zap.Logger
	kubernetesNamespace string
}

// NewCollector
func NewCollector(options CollectorOptions) (Collector, error) {

	gadgetCollectors := []GadgetCollector{
		{
			GadgetName: "dns",
			Callback: func(line string, node string) {
				var e dnstypes.Event

				if err := json.Unmarshal([]byte(line), &e); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s", err)
					return
				}

				baseEvent := e.GetBaseEvent()
				if baseEvent.Type != eventtypes.NORMAL {
					commonutils.ManageSpecialEvent(baseEvent, true)
					return
				}

				dnsResolutions.WithLabelValues(e.Node, string(e.Qr), e.Nameserver, e.DNSName).Inc()
			},
		},
		{
			GadgetName: "tcptracer",
			Callback: func(line string, node string) {
				var e tcptypes.Event

				if err := json.Unmarshal([]byte(line), &e); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s", err)
					return
				}

				baseEvent := e.GetBaseEvent()
				if baseEvent.Type != eventtypes.NORMAL {
					commonutils.ManageSpecialEvent(baseEvent, true)
					return
				}

				tcpResolutions.WithLabelValues(e.Node,
					e.Namespace,
					e.Pod,
					e.Container,
					e.Operation,
					e.Saddr,
					e.Daddr,
					strconv.Itoa(int(e.Dport))).Inc()
			},
		},
	}

	return &collector{
		gadgetCollectors:    gadgetCollectors,
		logger:              options.Logger,
		kubernetesNamespace: options.KubernetesNamespace}, nil
}

// Collect
func (c *collector) Collect(ctx context.Context) error {

	var wg sync.WaitGroup

	for _, gc := range c.gadgetCollectors {

		var commonFlags utils.CommonFlags

		if c.kubernetesNamespace == "" {
			commonFlags.AllNamespaces = true
		} else {
			commonFlags.Namespace = c.kubernetesNamespace
		}

		config := &utils.TraceConfig{
			GadgetName:       gc.GadgetName,
			Operation:        gadgetv1alpha1.OperationStart,
			TraceOutputMode:  gadgetv1alpha1.TraceOutputModeStream,
			TraceOutputState: gadgetv1alpha1.TraceStateStarted,
			CommonFlags:      &commonFlags,
			Parameters:       map[string]string{},
		}

		c.logger.Sugar().Infow("Adding Gadget...", "Gadget", gc.GadgetName)

		cb := gc.Callback
		wg.Add(1)
		go func() {
			err := utils.RunTraceStreamCallback(config, cb)
			if err != nil {
				c.logger.Sugar().Error(err)
			}
			defer wg.Done()
		}()
	}

	wg.Wait()

	return nil
}
