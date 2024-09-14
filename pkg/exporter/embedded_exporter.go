package exporter

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gizmoguy/exabgp_exporter/pkg/exabgp"
)

type EmbeddedExporter struct {
	mutex   sync.RWMutex
	summary *prometheus.GaugeVec
	rib     *prometheus.GaugeVec
	BaseExporter
}

func NewEmbeddedExporter(logger *slog.Logger) (*EmbeddedExporter, error) {
	be := NewBaseExporter(logger)
	be.up.Set(float64(1))

	sm := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer",
		Namespace: namespace,
		Subsystem: "state",
		Help:      summaryHelp,
	}, summaryLabelNames)
	rm := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "route",
		Namespace: namespace,
		Subsystem: "state",
		Help:      ribHelp,
	}, ribLabelNames)

	prometheus.MustRegister(sm)
	prometheus.MustRegister(rm)
	return &EmbeddedExporter{
		summary:      sm,
		rib:          rm,
		BaseExporter: be,
	}, nil
}

// Run starts the background reader for populating metrics
func (e *EmbeddedExporter) Run(reader *bufio.Reader) {
	go func() {
		for {
			line, _, err := reader.ReadLine()
			if err != nil && err != io.EOF {
				e.BaseExporter.logger.Error("unknown error", "error", err.Error())
				e.BaseExporter.parseFailures.Inc()
				continue
			}
			evt, err := exabgp.ParseEvent(line)
			if err != nil {
				e.BaseExporter.logger.Error(
					"unable to parse line", "error", err.Error(), "line", line,
				)
				e.BaseExporter.parseFailures.Inc()
				continue
			}
			var labels = map[string]string{
				"peer_ip":  evt.Peer.IP,
				"peer_asn": fmt.Sprintf("%d", evt.Peer.ASN),
			}
			switch evt.Peer.State {
			case "down":
				e.summary.With(labels).Set(float64(0))
			default:
				e.summary.With(labels).Set(float64(1))
			}
			if evt.Direction == "send" {
				announcements := evt.GetAnnouncements()
				if announcements != nil {
					labels["local_ip"] = evt.Self.IP
					labels["local_asn"] = fmt.Sprintf("%d", evt.Self.ASN)
					for _, v := range announcements.IPV4Unicast {
						labels["communities"] = communityToString(v.Attributes.Community)
						labels["as_path"] = asPathToString(v.Attributes.ASPath)
						labels["local_preference"] = strconv.Itoa(v.Attributes.LocalPreference)
						labels["med"] = strconv.Itoa(int(v.Attributes.Med))
						labels["family"] = "ipv4 unicast"
						for _, r := range v.NLRI {
							labels["nlri"] = r
							e.rib.With(labels).Set(float64(1))
						}
					}
					for _, v := range announcements.IPV6Unicast {
						labels["communities"] = communityToString(v.Attributes.Community)
						labels["as_path"] = asPathToString(v.Attributes.ASPath)
						labels["local_preference"] = strconv.Itoa(v.Attributes.LocalPreference)
						labels["med"] = strconv.Itoa(int(v.Attributes.Med))
						labels["family"] = "ipv6 unicast"
						for _, r := range v.NLRI {
							labels["nlri"] = r
							e.rib.With(labels).Set(float64(1))
						}
					}
				}
				withdraws := evt.GetWithdrawals()
				if withdraws != nil {
					labels["local_ip"] = evt.Self.IP
					labels["local_asn"] = fmt.Sprintf("%d", evt.Self.ASN)
					for _, w := range withdraws.IPv4Unicast {
						labels["communities"] = communityToString(w.Attributes.Community)
						labels["as_path"] = asPathToString(w.Attributes.ASPath)
						labels["local_preference"] = strconv.Itoa(w.Attributes.LocalPreference)
						labels["med"] = strconv.Itoa(int(w.Attributes.Med))
						for _, r := range w.NLRI {
							labels["family"] = "ipv4 unicast"
							labels["nlri"] = r
							e.rib.With(labels).Set(float64(0))
						}
					}
					for _, w := range withdraws.IPv6Unicast {
						labels["communities"] = communityToString(w.Attributes.Community)
						labels["as_path"] = asPathToString(w.Attributes.ASPath)
						labels["local_preference"] = strconv.Itoa(w.Attributes.LocalPreference)
						labels["med"] = strconv.Itoa(int(w.Attributes.Med))
						for _, r := range w.NLRI {
							labels["family"] = "ipv6 unicast"
							labels["nlri"] = r
							e.rib.With(labels).Set(float64(0))
						}
					}
				}
			}
		}
	}()
}

// Collect delivers all seen stats as Prometheus metrics
// It implements prometheus.Collector.
func (e *EmbeddedExporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.BaseExporter.totalScrapes.Inc()
	ch <- e.BaseExporter.totalScrapes
	ch <- e.BaseExporter.parseFailures
	ch <- e.BaseExporter.up
}

// Describe describes all the metrics ever exported by the exabgp exporter
// It implements prometheus.Collector
func (e *EmbeddedExporter) Describe(ch chan<- *prometheus.Desc) {
	e.BaseExporter.Describe(ch)
}

// Transform communities to string
func communityToString(communityAttribute [][]int) string {
	communityStrings := []string{}
	for _, community := range communityAttribute {
		// #0 is ASN and #1 is community value
		communityString := fmt.Sprintf("%d:%d", community[0], community[1])
		communityStrings = append(communityStrings, communityString)
	}

	return strings.Join(communityStrings, " ")
}

// Transform ASPath to string
func asPathToString(asPathAttribute []int) string {
	asPathStrings := []string{}
	for _, communityAS := range asPathAttribute {
		asPathStrings = append(asPathStrings, strconv.Itoa(communityAS))
	}

	return strings.Join(asPathStrings, " ")
}
