package config

import (
	"os"
	"strconv"
)

type Config struct {
	ListenAddr         string
	AdminToken         string
	DataDir            string
	ClusterName        string
	ManagedNamespace   string
	AllowAllNamespaces bool
}

func FromEnv() Config {
	cfg := Config{
		ListenAddr:       env("LISTEN_ADDR", ":8080"),
		AdminToken:       env("ADMIN_TOKEN", ""),
		DataDir:          env("DATA_DIR", "/data"),
		ClusterName:      env("CLUSTER_NAME", ""),
		ManagedNamespace: env("MANAGED_NAMESPACE", env("POD_NAMESPACE", "default")),
	}

	allowAll, _ := strconv.ParseBool(env("ALLOW_ALL_NAMESPACES", "false"))
	cfg.AllowAllNamespaces = allowAll
	return cfg
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
