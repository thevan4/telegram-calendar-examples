package main

import (
	"log"
	"os"
	"time"
)

var (
	grpcPort          string
	httpPort          string
	grpcDialTimeout   time.Duration
	httpDialTimeout   time.Duration
	secretBearerToken = ""
)

func init() {
	log.SetFlags(0)
	grpcPort = os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
		logJSON("debug", "grpc port set default: 55555")
	}

	httpPort = os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
		logJSON("debug", "http port set default: 55555")
	}

	var err error
	grpcDialTimeoutRaw := os.Getenv("GRPC_DIAL_TIMEOUT")
	if grpcDialTimeoutRaw == "" {
		grpcDialTimeout = time.Second
		logJSON("debug", "grpc dial timeout set default: 1s")
	} else {
		grpcDialTimeout, err = time.ParseDuration(grpcDialTimeoutRaw)
		if err != nil {
			grpcDialTimeout = time.Second
			logJSON("warn", "grpc dial timeout set error: "+err.Error()+", set default: 1s")
		} else {
			logJSON("debug", "grpc dial timeout set: "+grpcDialTimeout.String())
		}
	}

	httpDialTimeoutRaw := os.Getenv("HTTP_DIAL_TIMEOUT")
	if httpDialTimeoutRaw == "" {
		httpDialTimeout = time.Second
		logJSON("debug", "http dial timeout set default: 1s")
	} else {
		httpDialTimeout, err = time.ParseDuration(httpDialTimeoutRaw)
		if err != nil {
			httpDialTimeout = time.Second
			logJSON("warn", "http dial timeout set error: "+err.Error()+", set default: 1s")
		} else {
			logJSON("debug", "http dial timeout set: "+httpDialTimeout.String())
		}
	}

	secretBearerToken = os.Getenv("BEARER_TOKEN")
	if secretBearerToken == "" {
		logJSON("warn", "got empty secret bearer token, no auth at server now")
	}
}
