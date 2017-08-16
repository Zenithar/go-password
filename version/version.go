package version

import "github.com/prometheus/client_golang/prometheus"

// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion string
)

// Map provides the iterable version information.
var Map = map[string]string{
	"version":   Version,
	"revision":  Revision,
	"branch":    Branch,
	"buildUser": BuildUser,
	"buildDate": BuildDate,
	"goVersion": GoVersion,
}

func init() {
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "password_svc_build_info",
			Help: "A metric with a constant '1' value labeled by version, revision, and branch from which password service was built.",
		},
		[]string{"version", "revision", "branch"},
	)
	buildInfo.WithLabelValues(Version, Revision, Branch).Set(1)

	prometheus.MustRegister(buildInfo)
}
