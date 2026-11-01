package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/filesystem"

	"goravel/app-modules/media/models"
	"goravel/app/facades"
)

// Owner is what a file belongs to: a short stable type ("user", "page", "library") and its id.
type Owner struct {
	Type string
	ID   string
}

// UserOwner is a person's files.
func UserOwner(userID uint) Owner {
	return Owner{Type: "user", ID: strconv.FormatUint(uint64(userID), 10)}
}

// Errors a caller can show to a person.
var (
	ErrUnknownCollection = errors.New("unknown media collection")
	ErrTooLarge          = errors.New("the file is too large")
	ErrWrongType         = errors.New("this kind of file is not accepted here")
)

// Disks: MEDIA_STORAGE=r2 keeps files in Cloudflare R2 (public and private buckets), otherwise on this server.
func diskFor(public bool) string {
	storage, _ := facades.Config().Env("MEDIA_STORAGE", "local").(string)
	switch {
	case storage == "r2" && public:
		return "r2_public"
	case storage == "r2":
		return "r2_private"
	case public:
		return "public"
	default:
		return "local"
	}
}

func prefix() string {
	value, _ := facades.Config().Env("MEDIA_PREFIX", "media").(string)
	return strings.Trim(value, "/")
}

// Path is where a file sits on its disk: <prefix>/<id>/<file_name>, like Spatie's default path generator.
func Path(item models.Media) string {
	return prefix() + "/" + strconv.FormatUint(uint64(item.ID), 10) + "/" + item.FileName
}

var unsafeName = regexp.MustCompile(`[^a-z0-9._-]+`)

// SafeFileName keeps letters, digits, dot, dash and underscore.
func SafeFileName(name string) string {
	name = unsafeName.ReplaceAllString(strings.ToLower(filepath.Base(name)), "-")
	name = strings.Trim(name, "-.")
	if name == "" {
		return "file"
	}
	return name
}

// Add stores an uploaded file in the owner's collection (Spatie's addMedia(...)->toMediaCollection(...)).
func Add(owner Owner, collectionName string, file filesystem.File) (models.Media, error) {
	collection, ok := Find(owner.Type, collectionName)
	if !ok {
		return models.Media{}, ErrUnknownCollection
	}
	size, err := file.Size()
	if err != nil {
		return models.Media{}, err
	}
	if size > collection.limit() {
		return models.Media{}, ErrTooLarge
	}
	mime, _ := file.MimeType()
	if !collection.accepts(mime) {
		return models.Media{}, ErrWrongType
	}
	content, err := os.ReadFile(file.File())
	if err != nil {
		return models.Media{}, err
	}
	return AddBytes(owner, collectionName, file.GetClientOriginalName(), mime, content)
}

// AddBytes stores content (a generated PDF, an imported file) in a collection.
func AddBytes(owner Owner, collectionName, originalName, mime string, content []byte) (models.Media, error) {
	collection, ok := Find(owner.Type, collectionName)
	if !ok {
		return models.Media{}, ErrUnknownCollection
	}
	if int64(len(content)) > collection.limit() {
		return models.Media{}, ErrTooLarge
	}
	if !collection.accepts(mime) {
		return models.Media{}, ErrWrongType
	}
	var last models.Media
	_ = facades.Orm().Query().Where("model_type", owner.Type).Where("model_id", owner.ID).
		Where("collection_name", collectionName).OrderBy("order_column", "desc").First(&last)
	order := uint(1)
	if last.OrderColumn != nil {
		order = *last.OrderColumn + 1
	}
	properties, _ := json.Marshal(map[string]any{"public": collection.Public})
	id := uuid.NewString()
	fileName := SafeFileName(originalName)
	item := models.Media{
		ModelType: owner.Type, ModelID: owner.ID, UUID: &id, CollectionName: collectionName,
		Name: strings.TrimSuffix(fileName, filepath.Ext(fileName)), FileName: fileName, MimeType: &mime,
		Disk: diskFor(collection.Public), Size: uint64(len(content)), Manipulations: "[]",
		CustomProperties: string(properties), ResponsiveImages: "[]", OrderColumn: &order,
	}
	empty := "[]"
	item.GeneratedConversions = &empty
	// Like Spatie, the row comes first: the file's folder is the row's id.
	if err := facades.Orm().Query().Create(&item); err != nil {
		return models.Media{}, err
	}
	if err := facades.Storage().Disk(item.Disk).Put(Path(item), string(content)); err != nil {
		_, _ = facades.Orm().Query().Delete(&item)
		return models.Media{}, fmt.Errorf("storing the file failed: %w", err)
	}
	if collection.Single {
		trimTo(owner, collectionName, 1)
	} else if collection.Keep > 0 {
		trimTo(owner, collectionName, collection.Keep)
	}
	return item, nil
}

// trimTo keeps the newest count files of a collection and deletes the rest.
func trimTo(owner Owner, collectionName string, count int) {
	var items []models.Media
	_ = facades.Orm().Query().Where("model_type", owner.Type).Where("model_id", owner.ID).
		Where("collection_name", collectionName).OrderBy("order_column", "desc").OrderBy("id", "desc").Find(&items)
	for index, item := range items {
		if index >= count {
			_ = Delete(item)
		}
	}
}

// Delete removes the file and its row.
func Delete(item models.Media) error {
	_ = facades.Storage().Disk(item.Disk).Delete(Path(item))
	_, err := facades.Orm().Query().Delete(&item)
	return err
}

// Get is the files of an owner's collection, oldest first.
func Get(owner Owner, collectionName string) []models.Media {
	var items []models.Media
	_ = facades.Orm().Query().Where("model_type", owner.Type).Where("model_id", owner.ID).
		Where("collection_name", collectionName).OrderBy("order_column").Find(&items)
	return items
}

// First is the newest file of a collection, or nil.
func First(owner Owner, collectionName string) *models.Media {
	items := Get(owner, collectionName)
	if len(items) == 0 {
		return nil
	}
	return &items[len(items)-1]
}

// URL is the address of a file: the R2 public address for public files in R2, otherwise /media/{id}/{file_name}
// (served by this app, which checks the permission for private files).
func URL(item models.Media) string {
	if item.Disk == "r2_public" {
		return facades.Storage().Disk(item.Disk).Url(Path(item))
	}
	return "/media/" + strconv.FormatUint(uint64(item.ID), 10) + "/" + item.FileName
}
