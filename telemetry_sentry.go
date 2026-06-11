package main

import (
	"time"

	"github.com/getsentry/sentry-go"
)

const sentryFlushTimeout = 2 * time.Second

// PushToSentry reports a TGF error event to Sentry.
// Only called when exit_code != 0 and SentryDSN is configured.
// Errors during Sentry reporting are logged at debug level and never block execution.
func PushToSentry(cfg TelemetryConfig, event TGFEvent) {
	if cfg.SentryDSN == "" {
		return
	}

	if event.ExitCode == 0 {
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:     cfg.SentryDSN,
		Release: event.Version,
	})
	if err != nil {
		log.Debugf("Sentry: init failed: %v", err)
		return
	}
	defer sentry.Flush(sentryFlushTimeout)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("tgf_version", event.Version)
		scope.SetTag("image_version", event.ImageVersion)
		scope.SetTag("entry_point", event.EntryPoint)
		scope.SetTag("os", event.OS)
		scope.SetTag("arch", event.Arch)

		scope.SetExtra("image", event.Image)
		scope.SetExtra("exit_code", event.ExitCode)
		scope.SetExtra("duration_seconds", event.Duration)
		scope.SetExtra("hostname", event.Hostname)

		if event.Extra != nil {
			for k, v := range event.Extra {
				scope.SetExtra(k, v)
			}
		}
	})

	sentry.CaptureMessage(event.Error)
	log.Debugf("Sentry: reported error for image_version=%s exit_code=%d", event.ImageVersion, event.ExitCode)
}
