// Package backup dumps the app's database (mysqldump for MySQL/MariaDB, pg_dump for PostgreSQL), gzips it and
// stores it on the backup disk: <prefix>/<name>/<YYYY-MM-DD-HH-MM-SS>.sql.gz. Like Spatie's backup:run --only-db.
package backup

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"goravel/app/facades"
)

// Stored is one backup on the disk.
type Stored struct {
	Disk string
	Path string
	Size int64
}

// connection is the database the app uses, read from config/database.go.
type connection struct {
	driver   string
	host     string
	port     string
	database string
	username string
	password string
}

func appConnection() connection {
	config := facades.Config()
	name := config.GetString("database.default")
	key := "database.connections." + name + "."
	return connection{
		driver:   name,
		host:     config.GetString(key + "host"),
		port:     config.GetString(key + "port"),
		database: config.GetString(key + "database"),
		username: config.GetString(key + "username"),
		password: config.GetString(key + "password"),
	}
}

// Folder is where this app's backups live on the disk.
func Folder() string {
	config := facades.Config()
	return strings.Trim(config.GetString("backup.prefix"), "/") + "/" + strings.Trim(config.GetString("backup.name"), "/")
}

// Run dumps the database and stores it on the backup disk. A dump tool that fails or an empty dump stores nothing.
func Run(now time.Time) (Stored, error) {
	dump, err := dumpDatabase(appConnection())
	if err != nil {
		return Stored{}, err
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(dump); err != nil {
		return Stored{}, err
	}
	if err := writer.Close(); err != nil {
		return Stored{}, err
	}

	diskName := facades.Config().GetString("backup.disk")
	path := Folder() + "/" + now.UTC().Format("2006-01-02-15-04-05") + ".sql.gz"
	if err := facades.Storage().Disk(diskName).Put(path, compressed.String()); err != nil {
		return Stored{}, fmt.Errorf("upload to %s failed: %w", diskName, err)
	}
	return Stored{Disk: diskName, Path: path, Size: int64(compressed.Len())}, nil
}

// List returns the stored backups, oldest first.
func List() ([]string, error) {
	files, err := facades.Storage().Disk(facades.Config().GetString("backup.disk")).Files(Folder())
	if err != nil {
		return nil, err
	}
	backups := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, ".sql.gz") {
			// The disk lists names inside the folder; return full paths, the same shape Run returns.
			backups = append(backups, Folder()+"/"+strings.TrimPrefix(file, Folder()+"/"))
		}
	}
	sort.Strings(backups)
	return backups, nil
}

// Read downloads and unzips a stored backup (for restore checks and tests).
func Read(path string) (string, error) {
	content, err := facades.Storage().Disk(facades.Config().GetString("backup.disk")).GetBytes(path)
	if err != nil {
		return "", err
	}
	reader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		return "", err
	}
	sql, err := io.ReadAll(reader)
	return string(sql), err
}

func dumpDatabase(db connection) ([]byte, error) {
	var command *exec.Cmd
	switch db.driver {
	case "mysql":
		binary, err := dumpBinary("mysqldump", "mariadb-dump")
		if err != nil {
			return nil, err
		}
		args := []string{"--host", db.host, "--port", db.port, "--user", db.username, "--single-transaction",
			"--quick", "--skip-lock-tables", "--routines", "--triggers", "--default-character-set=utf8mb4"}
		if mysql8Client(binary) {
			// A MySQL 8 client asks a MariaDB server for a statistics table it does not have unless told not to.
			args = append(args, "--column-statistics=0")
		}
		command = exec.Command(binary, append(args, db.database)...)
		command.Env = append(os.Environ(), "MYSQL_PWD="+db.password)
	case "postgres":
		binary, err := dumpBinary("pg_dump")
		if err != nil {
			return nil, err
		}
		command = exec.Command(binary, "--host", db.host, "--port", db.port, "--username", db.username,
			"--no-password", "--no-owner", db.database)
		command.Env = append(os.Environ(), "PGPASSWORD="+db.password)
	default:
		return nil, fmt.Errorf("backups support mysql and postgres connections, not %q", db.driver)
	}

	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %s", command.Path, strings.TrimSpace(stderr.String()))
	}
	if len(output) == 0 {
		return nil, errors.New("the dump was empty")
	}
	return output, nil
}

// dumpBinary is the configured tool, or the first candidate that runs.
func dumpBinary(candidates ...string) (string, error) {
	if configured := facades.Config().GetString("backup.dump_binary"); configured != "" {
		candidates = []string{configured}
	}
	for _, candidate := range candidates {
		if exec.Command(candidate, "--version").Run() == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not run %s; install the database client tools or set BACKUP_DUMP_BINARY",
		strings.Join(candidates, " or "))
}

func mysql8Client(binary string) bool {
	version, _ := exec.Command(binary, "--version").Output()
	text := strings.ToLower(string(version))
	if strings.Contains(text, "mariadb") {
		return false
	}
	for _, major := range []string{"ver 8.", "ver 9."} {
		if strings.Contains(text, major) {
			return true
		}
	}
	return false
}
