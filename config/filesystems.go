package config

import (
	"strings"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/support/path"

	"goravel/app/facades"
	"goravel/app/storage"
)

// endpointHost turns "https://<account>.r2.cloudflarestorage.com" into the bare host the MinIO driver expects.
func endpointHost(value any) string {
	text, _ := value.(string)
	text = strings.TrimPrefix(strings.TrimPrefix(text, "https://"), "http://")
	return strings.TrimSuffix(text, "/")
}

// privateAddress is the bucket's S3 address. It needs the keys to read anything, so it is safe to put in config.
func privateAddress(endpoint, bucket any) string {
	host := endpointHost(endpoint)
	name, _ := bucket.(string)
	if host == "" {
		return ""
	}
	return "https://" + host + "/" + name
}

func init() {
	config := facades.Config()
	config.Add("filesystems", map[string]any{
		// Disk used by facades.Storage() without Disk(...): "local" in development, "r2_public" when uploads live in R2.
		"default": config.Env("FILESYSTEM_DISK", "local"),

		// Filesystem Disks
		//
		// Cloudflare R2 speaks the S3 API, so it uses the MinIO driver (it accepts a custom endpoint, which the S3
		// driver does not), wrapped by app/storage.R2 for correct public addresses. Two disks, never mixed: public uploads, and private database backups.
		//
		// Supported Drivers: "local", "custom"
		"disks": map[string]any{
			"local": map[string]any{
				"driver": "local",
				"root":   path.Storage("app"),
			},
			"public": map[string]any{
				"driver": "local",
				"root":   path.Storage("app/public"),
				"url":    config.Env("APP_URL", "").(string) + "/storage",
			},
			// Files anyone may see (served from R2_PUBLIC_URL).
			"r2_public": map[string]any{
				"driver": "custom",
				"key":    config.Env("R2_ACCESS_KEY_ID"),
				"secret": config.Env("R2_SECRET_ACCESS_KEY"),
				"region": config.Env("R2_REGION", "auto"),
				"bucket": config.Env("R2_PUBLIC_BUCKET"),
				// The MinIO driver needs a url; app/storage.R2 builds public addresses from public_url instead.
				"url":        privateAddress(config.Env("R2_ENDPOINT"), config.Env("R2_PUBLIC_BUCKET")),
				"public_url": config.Env("R2_PUBLIC_URL"),
				"endpoint":   endpointHost(config.Env("R2_ENDPOINT")),
				"ssl":        true,
				"via": func() (filesystem.Driver, error) {
					return storage.NewR2("r2_public")
				},
			},
			// Database backups: a private bucket with its own keys, never publicly readable.
			"r2_backups": map[string]any{
				"driver":     "custom",
				"key":        config.Env("BACKUP_ACCESS_KEY_ID"),
				"secret":     config.Env("BACKUP_SECRET_ACCESS_KEY"),
				"region":     config.Env("BACKUP_REGION", "auto"),
				"bucket":     config.Env("BACKUP_BUCKET"),
				"url":        privateAddress(config.Env("BACKUP_ENDPOINT"), config.Env("BACKUP_BUCKET")),
				"public_url": "", // private: no public address
				"endpoint":   endpointHost(config.Env("BACKUP_ENDPOINT")),
				"ssl":        true,
				"via": func() (filesystem.Driver, error) {
					return storage.NewR2("r2_backups")
				},
			},
		},
	})
}
