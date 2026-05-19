package main

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/alecthomas/kong"
	"github.com/go-logr/logr"
)

type managerOptions struct {
	MetricsBindAddress     string `name:"metrics-bind-address" default:"0" help:"Metrics bind address. Use 0 to disable."`
	HealthProbeBindAddress string `name:"health-probe-bind-address" default:":8081" help:"Health probe bind address."`
	LeaderElect            bool   `name:"leader-elect" default:"false" help:"Enable leader election."`
	MetricsSecure          bool   `name:"metrics-secure" default:"true" negatable:"" help:"Serve metrics over HTTPS."`
	WebhookCertPath        string `name:"webhook-cert-path" help:"Webhook certificate directory."`
	WebhookCertName        string `name:"webhook-cert-name" default:"tls.crt" help:"Webhook certificate filename."`
	WebhookCertKey         string `name:"webhook-cert-key" default:"tls.key" help:"Webhook private key filename."`
	MetricsCertPath        string `name:"metrics-cert-path" help:"Metrics certificate directory."`
	MetricsCertName        string `name:"metrics-cert-name" default:"tls.crt" help:"Metrics certificate filename."`
	MetricsCertKey         string `name:"metrics-cert-key" default:"tls.key" help:"Metrics private key filename."`
	EnableHTTP2            bool   `name:"enable-http2" default:"false" help:"Enable HTTP/2 for metrics and webhooks."`
	LogFormat              string `name:"log-format" enum:"json,text" default:"json" help:"Log output format."`
	LogLevel               string `name:"log-level" enum:"debug,info,warn,error" default:"info" help:"Minimum log level."`
}

func newManagerParser(options *managerOptions) (*kong.Kong, error) {
	return kong.New(options,
		kong.Name("manager"),
		kong.Description("Kubernetes controller manager."),
		kong.UsageOnError(),
	)
}

func parseManagerOptions(args []string) (managerOptions, error) {
	var options managerOptions
	parser, err := newManagerParser(&options)
	if err != nil {
		return managerOptions{}, err
	}

	if _, err := parser.Parse(args); err != nil {
		return managerOptions{}, err
	}

	return options, nil
}

func slogLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level %q", level)
	}
}

func newControllerLogger(options managerOptions, out io.Writer) (logr.Logger, error) {
	level, err := slogLevel(options.LogLevel)
	if err != nil {
		return logr.Logger{}, err
	}

	var levelVar slog.LevelVar
	levelVar.Set(level)

	handlerOptions := &slog.HandlerOptions{Level: &levelVar}
	switch options.LogFormat {
	case "json":
		return logr.FromSlogHandler(slog.NewJSONHandler(out, handlerOptions)), nil
	case "text":
		return logr.FromSlogHandler(slog.NewTextHandler(out, handlerOptions)), nil
	default:
		return logr.Logger{}, fmt.Errorf("unsupported log format %q", options.LogFormat)
	}
}
