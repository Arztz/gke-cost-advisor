package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Prometheus PrometheusConfig
	Kubernetes KubernetesConfig
	GCP        GCPConfig
	Analysis   AnalysisConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	Endpoint string
	Token    string
	Timeout  time.Duration
}

// KubernetesConfig holds Kubernetes configuration
type KubernetesConfig struct {
	Kubeconfig string
	Namespace  string
}

// GCPConfig holds Google Cloud configuration
type GCPConfig struct {
	ProjectID        string
	BillingAccountID string
	Region           string
}

// AnalysisConfig holds analysis configuration
type AnalysisConfig struct {
	DefaultWindow  time.Duration
	CPUHeadroom    float64
	MemoryHeadroom float64
	MaxWorkers     int
	QueryTimeout   time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		Prometheus: PrometheusConfig{
			Endpoint: getEnv("PROMETHEUS_ENDPOINT", "http://localhost:9090"),
			Token:    getEnv("PROMETHEUS_TOKEN", ""),
			Timeout:  getEnvDuration("PROMETHEUS_TIMEOUT", 30*time.Second),
		},
		Kubernetes: KubernetesConfig{
			Kubeconfig: getEnv("KUBECONFIG", ""),
			Namespace:  getEnv("KUBERNETES_NAMESPACE", ""),
		},
		GCP: GCPConfig{
			ProjectID:        getEnv("GCP_PROJECT_ID", ""),
			BillingAccountID: getEnv("GCP_BILLING_ACCOUNT_ID", ""),
			Region:           getEnv("GCP_REGION", "us-central1"),
		},
		Analysis: AnalysisConfig{
			DefaultWindow:  getEnvDuration("ANALYSIS_DEFAULT_WINDOW", 24*time.Hour),
			CPUHeadroom:    getEnvFloat("ANALYSIS_CPU_HEADROOM", 0.20),
			MemoryHeadroom: getEnvFloat("ANALYSIS_MEMORY_HEADROOM", 0.30),
			MaxWorkers:     getEnvInt("ANALYSIS_MAX_WORKERS", 10),
			QueryTimeout:   getEnvDuration("ANALYSIS_QUERY_TIMEOUT", 60*time.Second),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
