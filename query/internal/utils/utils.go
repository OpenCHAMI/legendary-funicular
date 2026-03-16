// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package utils
package utils

import (
	"fmt"
	"log/slog"
	"os"
)

func CheckFatal(err error) {
	if err != nil {
		slog.Error(fmt.Sprint(err))
		os.Exit(1)
	}
}

func GetEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return value, fmt.Errorf("undefined required environment variable: '%s'", name)
	}
	return value, nil
}

func GetEnvFatal(name string) string {
	value, err := GetEnv(name)
	CheckFatal(err)
	return value
}
