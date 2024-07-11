// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/log"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	opentelemetrySectionName string = "opentelemetry"
	alwaysOn                 string = "always_on"
	alwaysOff                string = "always_off"
	traceIDRatio             string = "traceidratio"
	parentBasedAlwaysOn      string = "parentbased_always_on"
	parentBasedAlwaysOff     string = "parentbased_always_off"
	parentBasedTraceIDRatio  string = "parentbased_traceidratio"
)

// Opentelemetry settings
var (
	OpenTelemetry = struct {
		Traces   traceConfig
		Resource resourceConfig
	}{
		Traces:   traceConfig{Timeout: 10 * time.Second},
		Resource: resourceConfig{ServiceName: "forgejo", EnabledDecoders: "all"},
	}
	compressions = []string{"gzip", ""}
)

type traceConfig struct {
	Endpoint          *url.URL // A base endpoint URL for any signal type, with an optionally-specified port number
	Headers           map[string]string
	Insecure          bool          // Disable TLS
	Compression       string        // Supported value - ""/"gzip"
	Timeout           time.Duration // The timeout value for all outgoing data
	Sampler           sdktrace.Sampler
	Certificate       string
	ClientKey         string
	ClientCertificate string
}
type resourceConfig struct {
	ServiceName     string // Value of the service.name resource attribute, defaults to APP_NAME in lowercase
	Attributes      string // unprocessed attributes for the resource
	EnabledDecoders string
}

func loadOpenTelemetryFrom(rootCfg ConfigProvider) {
	sec := rootCfg.Section(opentelemetrySectionName)
	loadResourceConfig(sec)
	loadTraceConfig(sec)
}

func loadResourceConfig(sec ConfigSection) {
	OpenTelemetry.Resource.ServiceName = sec.Key("SERVICE_NAME").MustString(OpenTelemetry.Resource.ServiceName)
	OpenTelemetry.Resource.Attributes = sec.Key("RESOURCE_ATTRIBUTES").String()
	OpenTelemetry.Resource.EnabledDecoders = sec.Key("ENABLE_DECODERS").MustString(OpenTelemetry.Resource.EnabledDecoders)
	OpenTelemetry.Resource.EnabledDecoders = strings.ToLower(strings.TrimSpace(OpenTelemetry.Resource.EnabledDecoders))
}

func loadTraceConfig(rootSec ConfigSection) {
	var endpoint string
	if rootSec.HasKey("EXPORTER_OTLP_TRACES_ENDPOINT") {
		endpoint = rootSec.Key("EXPORTER_OTLP_TRACES_ENDPOINT").String()
	} else {
		endpoint = rootSec.Key("EXPORTER_OTLP_ENDPOINT").String()
	}

	if endpoint == "" {
		return
	}

	if ep, err := url.Parse(endpoint); err == nil && ep.Host != "" {
		OpenTelemetry.Traces.Endpoint = ep
	} else if err != nil {
		log.Warn("Otel trace endpoint parsing failure, err=%s", err)
		return
	} else {
		log.Warn("Otel trace endpoint parsing failure, no host was detected")
		return
	}

	if OpenTelemetry.Traces.Endpoint.Scheme == "http" || OpenTelemetry.Traces.Endpoint.Scheme == "unix" {
		OpenTelemetry.Traces.Insecure = true
	}
	OpenTelemetry.Traces.Insecure = rootSec.Key("EXPORTER_OTLP_TRACES_INSECURE").MustBool(rootSec.Key("EXPORTER_OTLP_INSECURE").MustBool(OpenTelemetry.Traces.Insecure))

	OpenTelemetry.Traces.Compression = rootSec.Key("EXPORTER_OTLP_COMPRESSION").In(OpenTelemetry.Traces.Compression, compressions)
	if rootSec.HasKey("EXPORTER_OTLP_TRACES_COMPRESSION") {
		OpenTelemetry.Traces.Compression = rootSec.Key("EXPORTER_OTLP_TRACES_COMPRESSION").In(OpenTelemetry.Traces.Compression, compressions)
	}

	OpenTelemetry.Traces.Timeout = rootSec.Key("EXPORTER_OTLP_TRACES_TIMEOUT").MustDuration(rootSec.Key("EXPORTER_OTLP_TIMEOUT").MustDuration(OpenTelemetry.Traces.Timeout))
	samplers := make([]string, 0, len(sampler))
	for k := range sampler {
		samplers = append(samplers, k)
	}
	samplerName := rootSec.Key("TRACES_SAMPLER").In(parentBasedAlwaysOn, samplers)
	samplerArg := rootSec.Key("TRACES_SAMPLER_ARG").MustString("")
	OpenTelemetry.Traces.Sampler = sampler[samplerName](samplerArg)

	OpenTelemetry.Traces.Headers = map[string]string{}
	headers := rootSec.Key("EXPORTER_OTLP_HEADERS").String()
	if headers != "" {
		for k, v := range _stringToHeader(headers) {
			OpenTelemetry.Traces.Headers[k] = v
		}
	}
	headers = rootSec.Key("EXPORTER_OTLP_TRACES_HEADERS").String()
	if headers != "" {
		for k, v := range _stringToHeader(headers) {
			OpenTelemetry.Traces.Headers[k] = v
		}
	}

	OpenTelemetry.Traces.Certificate = rootSec.Key("EXPORTER_OTLP_TRACES_CERTIFICATE").MustString(rootSec.Key("EXPORTER_OTLP_CERTIFICATE").String())
	OpenTelemetry.Traces.ClientCertificate = rootSec.Key("EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE").MustString(rootSec.Key("EXPORTER_OTLP_CLIENT_CERTIFICATE").String())
	OpenTelemetry.Traces.ClientKey = rootSec.Key("EXPORTER_OTLP_TRACES_CLIENT_KEY").MustString(rootSec.Key("EXPORTER_OTLP_CLIENT_KEY").String())
	if len(OpenTelemetry.Traces.Certificate) > 0 && !filepath.IsAbs(OpenTelemetry.Traces.Certificate) {
		OpenTelemetry.Traces.Certificate = filepath.Join(CustomPath, OpenTelemetry.Traces.Certificate)
	}
	if len(OpenTelemetry.Traces.ClientCertificate) > 0 && !filepath.IsAbs(OpenTelemetry.Traces.ClientCertificate) {
		OpenTelemetry.Traces.ClientCertificate = filepath.Join(CustomPath, OpenTelemetry.Traces.ClientCertificate)
	}
	if len(OpenTelemetry.Traces.ClientKey) > 0 && !filepath.IsAbs(OpenTelemetry.Traces.ClientKey) {
		OpenTelemetry.Traces.ClientKey = filepath.Join(CustomPath, OpenTelemetry.Traces.ClientKey)
	}
}

var sampler = map[string]func(arg string) sdktrace.Sampler{
	alwaysOff: func(_ string) sdktrace.Sampler {
		return sdktrace.NeverSample()
	},
	alwaysOn: func(_ string) sdktrace.Sampler {
		return sdktrace.AlwaysSample()
	},
	traceIDRatio: func(arg string) sdktrace.Sampler {
		ratio, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			ratio = 1
		}
		return sdktrace.TraceIDRatioBased(ratio)
	},
	parentBasedAlwaysOff: func(_ string) sdktrace.Sampler {
		return sdktrace.ParentBased(sdktrace.NeverSample())
	},
	parentBasedAlwaysOn: func(_ string) sdktrace.Sampler {
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	},
	parentBasedTraceIDRatio: func(arg string) sdktrace.Sampler {
		ratio, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			ratio = 1
		}
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
	},
}

// Opentelemetry SDK function port

func _stringToHeader(value string) map[string]string {
	headersPairs := strings.Split(value, ",")
	headers := make(map[string]string)

	for _, header := range headersPairs {
		n, v, found := strings.Cut(header, "=")
		if !found {
			log.Warn("Otel header ignored, err=\"missing '='\", input=%s", header)
			continue
		}
		name, err := url.PathUnescape(n)
		if err != nil {
			log.Warn("Otel header ignored, err=\"escape header key\", key=%s", n)
			continue
		}
		trimmedName := strings.TrimSpace(name)
		value, err := url.PathUnescape(v)
		if err != nil {
			log.Warn("Otel header ignored, err=\"escape header value\", value=%s", v)
			continue
		}
		trimmedValue := strings.TrimSpace(value)

		headers[trimmedName] = trimmedValue
	}

	return headers
}

func IsOpenTelemetryEnabled() bool {
	return OpenTelemetry.Traces.Endpoint != nil
}
