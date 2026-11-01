// Package storage holds this app's filesystem drivers.
package storage

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/filesystem"
	miniofacades "github.com/goravel/minio/facades"

	"goravel/app/facades"
)

// R2 is a Cloudflare R2 disk. It uses the official MinIO driver for every operation, and only fixes Url():
// MinIO builds path-style addresses (https://host/<bucket>/<file>), but an R2 public address (r2.dev or a custom
// domain) is bound to one bucket and serves https://host/<file>.
type R2 struct {
	filesystem.Driver
	publicURL string
}

// NewR2 opens the disk named `disk` in config/filesystems.go.
func NewR2(disk string) (filesystem.Driver, error) {
	driver, err := miniofacades.Minio(disk)
	if err != nil {
		return nil, err
	}
	return &R2{Driver: driver, publicURL: facades.Config().GetString("filesystems.disks." + disk + ".public_url")}, nil
}

// Url is the public address of the file. A disk without a public URL (a private bucket) returns "".
func (r *R2) Url(file string) string {
	if r.publicURL == "" {
		return ""
	}
	return strings.TrimSuffix(r.publicURL, "/") + "/" + strings.TrimPrefix(file, "/")
}

// WithContext keeps the R2 wrapper, so Url stays correct on a disk used with a context.
func (r *R2) WithContext(ctx context.Context) filesystem.Driver {
	return &R2{Driver: r.Driver.WithContext(ctx), publicURL: r.publicURL}
}
