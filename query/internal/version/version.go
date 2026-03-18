// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
//
// SPDX-FileNotice: This file is derived from:
// https://github.com/OpenCHAMI/ochami/blob/2247677c4e2a944e230872b5024100078dd8efec/internal/version/version.go
// Changes: adjusted import paths for local module usage; minor adjustments.

// Package version
package version

const ProgName = "openchami-logq"

var (
	// Basic values
	Version = "unknown"
	Tag     = "unknown"
	Branch  = "unknown"
	Commit  = "unknown"
	Date    = "unknown"

	// Other useful values
	GoVersion = "unknown"
	GitState  = "unknown"
	BuildHost = "unknown"
	BuildUser = "unknown"
)
