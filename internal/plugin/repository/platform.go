// Package repository provides plugin repository management for Tinct.
package repository

import "fmt"

// NormalizePlatform converts Go's GOOS/GOARCH to repository platform naming.
// Repository uses "x86" instead of "amd64" for compatibility with other languages.
func NormalizePlatform(goos, goarch string) string {
	arch := goarch
	switch goarch {
	case "amd64":
		arch = "x86"
	case "386":
		arch = "x86"
	}
	return fmt.Sprintf("%s_%s", goos, arch)
}
