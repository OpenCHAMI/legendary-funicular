// SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

// Package dev
package dev

import "errors"

func NotImplemented() error {
	return errors.New("not implemented")
}
