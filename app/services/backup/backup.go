// Package backup dumps the app's database (mysqldump for MySQL/MariaDB, pg_dump for PostgreSQL), gzips it and
// stores it on the backup disk: <prefix>/<name>/<YYYY-MM-DD-HH-MM-SS>.sql.gz. Like Spatie's backup:run --only-db.
//
// Nothing is held in memory: the dump streams through gzip into a temporary file in storage/app/backup-temp (like
// Spatie's backup-temp), then goes to R2 as a multipart upload one 16 MB part at a time. A database of any size
// needs only a few megabytes of memory and free disk space for the compressed file.
package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	supportpath "github.com/goravel/framework/support/path"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

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

// Run dumps the database and stores it on the backup disk. A dump tool that fails or an empty dump stores nothing,
// and the temporary file is always removed.
func Run(now time.Time) (Stored, error) {
	tempDir := TemporaryDirectory()
	if err := os.MkdirAll(tempDir, 0o700); err != nil {
		return Stored{}, err
	}
	removeStaleFiles(tempDir, 24*time.Hour)
	tempFile := filepath.Join(tempDir, now.UTC().Format("2006-01-02-15-04-05")+"-"+strconv.Itoa(os.Getpid())+".sql.gz")
	defer os.Remove(tempFile)

	if err := dumpToGzipFile(appConnection(), tempFile); err != nil {
		return Stored{}, err
	}
	info, err := os.Stat(tempFile)
	if err != nil {
		return Stored{}, err
	}
	diskName := facades.Config().GetString("backup.disk")
	path := Folder() + "/" + now.UTC().Format("2006-01-02-15-04-05") + ".sql.gz"
	if err := upload(diskName, path, tempFile); err != nil {
		return Stored{}, fmt.Errorf("upload to %s failed: %w", diskName, err)
	}
	return Stored{Disk: diskName, Path: path, Size: info.Size()}, nil
}

// TemporaryDirectory is BACKUP_TEMPORARY_DIRECTORY, default storage/app/backup-temp.
func TemporaryDirectory() string {
	if dir := facades.Config().GetString("backup.temporary_directory"); dir != "" {
		return dir
	}
	return supportpath.Storage("app/backup-temp")
}

// removeStaleFiles deletes what a crashed run left behind.
func removeStaleFiles(dir string, olderThan time.Duration) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if info, err := entry.Info(); err == nil && !entry.IsDir() && time.Since(info.ModTime()) > olderThan {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}

// upload sends the file to the disk without loading it: R2 and other S3 disks get a multipart upload, one part
// in memory at a time; a local disk gets a streamed copy.
func upload(diskName, objectPath, file string) error {
	key := "filesystems.disks." + diskName + "."
	config := facades.Config()
	if config.GetString(key+"driver") == "local" {
		target := filepath.Join(config.GetString(key+"root"), filepath.FromSlash(objectPath))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		return copyFile(file, target)
	}
	client, err := minio.New(config.GetString(key+"endpoint"), &minio.Options{
		Creds:  credentials.NewStaticV4(config.GetString(key+"key"), config.GetString(key+"secret"), ""),
		Secure: config.GetBool(key+"ssl", true),
		Region: config.GetString(key + "region"),
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()
	_, err = client.FPutObject(ctx, config.GetString(key+"bucket"), objectPath, file, minio.PutObjectOptions{
		ContentType: "application/gzip",
		PartSize:    PartSize,
		NumThreads:  1, // one part in memory at a time
	})
	return err
}

// PartSize is the size of one upload part (S3 allows at most 10,000 parts: 16 MB parts cover 160 GB).
const PartSize = 16 << 20

func copyFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
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

// dumpCommand builds the dump tool's command for the connection. The password goes in the environment.
func dumpCommand(db connection) (*exec.Cmd, error) {
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
		command := exec.Command(binary, append(args, db.database)...)
		command.Env = append(os.Environ(), "MYSQL_PWD="+db.password)
		return command, nil
	case "postgres":
		binary, err := dumpBinary("pg_dump")
		if err != nil {
			return nil, err
		}
		command := exec.Command(binary, "--host", db.host, "--port", db.port, "--username", db.username,
			"--no-password", "--no-owner", db.database)
		command.Env = append(os.Environ(), "PGPASSWORD="+db.password)
		return command, nil
	default:
		return nil, fmt.Errorf("backups support mysql and postgres connections, not %q", db.driver)
	}
}

// dumpToGzipFile streams the dump tool's output through gzip into target, 32 KB at a time. It fails, and leaves
// no usable file, when the tool fails or writes nothing.
func dumpToGzipFile(db connection, target string) error {
	command, err := dumpCommand(db)
	if err != nil {
		return err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr limitedBuffer
	command.Stderr = &stderr
	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := command.Start(); err != nil {
		return fmt.Errorf("could not start %s: %w", command.Path, err)
	}
	compressor := gzip.NewWriter(file)
	written, copyErr := io.Copy(compressor, stdout)
	closeErr := compressor.Close()
	waitErr := command.Wait()
	switch {
	case waitErr != nil:
		return fmt.Errorf("%s failed: %s", command.Path, strings.TrimSpace(stderr.String()))
	case copyErr != nil:
		return copyErr
	case closeErr != nil:
		return closeErr
	case written == 0:
		return errors.New("the dump was empty")
	}
	return file.Sync()
}

// limitedBuffer keeps the first 8 KB of the tool's error output, enough to say what went wrong.
type limitedBuffer struct{ bytes.Buffer }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if room := 8192 - b.Len(); room > 0 {
		if len(p) > room {
			b.Buffer.Write(p[:room])
		} else {
			b.Buffer.Write(p)
		}
	}
	return len(p), nil
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
