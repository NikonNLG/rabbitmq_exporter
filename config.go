package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	config        rabbitExporterConfig
	defaultConfig = rabbitExporterConfig{
		RabbitURL:          "http://127.0.0.1:15672",
		RabbitUsername:     "guest",
		RabbitPassword:     "guest",
		PublishPort:        "9419",
		PublishAddr:        "",
		OutputFormat:       "TTY", //JSON
		CAFile:             "ca.pem",
		InsecureSkipVerify: false,
		ExcludeMetrics:     []string{},
		SkipQueues:         regexp.MustCompile("^$"),
		IncludeQueues:      regexp.MustCompile(".*"),
		SkipVHost:          regexp.MustCompile("^$"),
		IncludeVHost:       regexp.MustCompile(".*"),
		RabbitCapabilities: parseCapabilities("no_sort,bert"),
		EnabledExporters:   []string{"exchange", "node", "overview", "queue"},
		Timeout:            30,
		MaxQueues:          0,
	}
)

type rabbitExporterConfig struct {
	RabbitURL          string
	RabbitUsername     string
	RabbitPassword     string
	PublishPort        string
	PublishAddr        string
	OutputFormat       string
	CAFile             string
	InsecureSkipVerify bool
	ExcludeMetrics     []string
	SkipQueues         *regexp.Regexp
	IncludeQueues      *regexp.Regexp
	SkipVHost          *regexp.Regexp
	IncludeVHost       *regexp.Regexp
	RabbitCapabilities rabbitCapabilitySet
	EnabledExporters   []string
	Timeout            int
	MaxQueues          int
}

type rabbitCapability string
type rabbitCapabilitySet map[rabbitCapability]bool

const (
	rabbitCapNoSort rabbitCapability = "no_sort"
	rabbitCapBert   rabbitCapability = "bert"
)

var allRabbitCapabilities = rabbitCapabilitySet{
	rabbitCapNoSort: true,
	rabbitCapBert:   true,
}

func initConfig() {
	config = defaultConfig
	if url := os.Getenv("RABBIT_URL"); url != "" {
		if valid, _ := regexp.MatchString("https?://[a-zA-Z.0-9]+", strings.ToLower(url)); valid {
			config.RabbitURL = url
		} else {
			panic(fmt.Errorf("Rabbit URL must start with http:// or https://"))
		}
	}

	var user string
	var pass string

	if len(os.Getenv("RABBIT_USER_FILE")) != 0 {
		fileContents, err := ioutil.ReadFile(os.Getenv("RABBIT_USER_FILE"))
		if err != nil {
			panic(err)
		}
		user = strings.TrimSpace(string(fileContents))
	} else {
		user = os.Getenv("RABBIT_USER")
	}

	if user != "" {
		config.RabbitUsername = user
	}

	if len(os.Getenv("RABBIT_PASSWORD_FILE")) != 0 {
		fileContents, err := ioutil.ReadFile(os.Getenv("RABBIT_PASSWORD_FILE"))
		if err != nil {
			panic(err)
		}
		pass = strings.TrimSpace(string(fileContents))
	} else {
		pass = os.Getenv("RABBIT_PASSWORD")
	}
	if pass != "" {
		config.RabbitPassword = pass
	}

	if port := os.Getenv("PUBLISH_PORT"); port != "" {
		if _, err := strconv.Atoi(port); err == nil {
			config.PublishPort = port
		} else {
			panic(fmt.Errorf("The configured port is not a valid number: %v", port))
		}

	}

	if addr := os.Getenv("PUBLISH_ADDR"); addr != "" {
		config.PublishAddr = addr
	}

	if output := os.Getenv("OUTPUT_FORMAT"); output != "" {
		config.OutputFormat = output
	}

	if cafile := os.Getenv("CAFILE"); cafile != "" {
		config.CAFile = cafile
	}
	if insecureSkipVerify := os.Getenv("SKIPVERIFY"); insecureSkipVerify == "true" || insecureSkipVerify == "1" || insecureSkipVerify == "TRUE" {
		config.InsecureSkipVerify = true
	}

	if ExcludeMetrics := os.Getenv("EXCLUDE_METRICS"); ExcludeMetrics != "" {
		config.ExcludeMetrics = strings.Split(ExcludeMetrics, ",")
	}

	if SkipQueues := os.Getenv("SKIP_QUEUES"); SkipQueues != "" {
		config.SkipQueues = regexp.MustCompile(SkipQueues)
	}

	if IncludeQueues := os.Getenv("INCLUDE_QUEUES"); IncludeQueues != "" {
		config.IncludeQueues = regexp.MustCompile(IncludeQueues)
	}

	if SkipVHost := os.Getenv("SKIP_VHOST"); SkipVHost != "" {
		config.SkipVHost = regexp.MustCompile(SkipVHost)
	}

	if IncludeVHost := os.Getenv("INCLUDE_VHOST"); IncludeVHost != "" {
		config.IncludeVHost = regexp.MustCompile(IncludeVHost)
	}

	if rawCapabilities := os.Getenv("RABBIT_CAPABILITIES"); rawCapabilities != "" {
		config.RabbitCapabilities = parseCapabilities(rawCapabilities)
	}

	if enabledExporters := os.Getenv("RABBIT_EXPORTERS"); enabledExporters != "" {
		config.EnabledExporters = strings.Split(enabledExporters, ",")
	}

	if timeout := os.Getenv("RABBIT_TIMEOUT"); timeout != "" {
		t, err := strconv.Atoi(timeout)
		if err != nil {
			panic(fmt.Errorf("timeout is not a number: %v", err))
		}
		config.Timeout = t
	}

	if maxQueues := os.Getenv("MAX_QUEUES"); maxQueues != "" {
		m, err := strconv.Atoi(maxQueues)
		if err != nil {
			panic(fmt.Errorf("maxQueues is not a number: %v", err))
		}
		config.MaxQueues = m
	}
}

func parseCapabilities(raw string) rabbitCapabilitySet {
	result := make(rabbitCapabilitySet)
	candidates := strings.Split(raw, ",")
	for _, maybeCapStr := range candidates {
		maybeCap := rabbitCapability(strings.TrimSpace(maybeCapStr))
		enabled, present := allRabbitCapabilities[maybeCap]
		if enabled && present {
			result[maybeCap] = true
		}
	}
	return result
}

func isCapEnabled(config rabbitExporterConfig, cap rabbitCapability) bool {
	exists, enabled := config.RabbitCapabilities[cap]
	return exists && enabled
}
