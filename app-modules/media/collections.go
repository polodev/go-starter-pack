package media

import (
	"strings"
	"sync"
)

// DefaultMaxBytes is the size limit of a collection that sets none (10 MB).
const DefaultMaxBytes = 10 << 20

// Collection is a named group of files of one owner type, like Spatie's addMediaCollection.
type Collection struct {
	Name     string
	Single   bool     // a new file replaces the old one
	Keep     int      // multiple: keep the newest N (0 = all)
	MaxBytes int64    // 0 = DefaultMaxBytes
	Accepts  []string // mime types or prefixes ("image/"); empty = anything
	Public   bool     // anyone may see the file; otherwise owner and staff with media.view only
}

// Single makes a one-file collection.
func Single(name string) Collection { return Collection{Name: name, Single: true} }

// Multiple makes a many-files collection.
func Multiple(name string) Collection { return Collection{Name: name} }

func (c Collection) Images() Collection {
	c.Accepts = []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	return c
}
func (c Collection) AsPublic() Collection            { c.Public = true; return c }
func (c Collection) KeepNewest(count int) Collection { c.Keep = count; return c }
func (c Collection) Max(bytes int64) Collection      { c.MaxBytes = bytes; return c }

func (c Collection) limit() int64 {
	if c.MaxBytes > 0 {
		return c.MaxBytes
	}
	return DefaultMaxBytes
}

func (c Collection) accepts(mime string) bool {
	if len(c.Accepts) == 0 {
		return true
	}
	for _, accepted := range c.Accepts {
		if mime == accepted || (strings.HasSuffix(accepted, "/") && strings.HasPrefix(mime, accepted)) {
			return true
		}
	}
	return false
}

var (
	collectionsMutex sync.RWMutex
	collections      = map[string]map[string]Collection{} // owner type -> name -> collection
)

// Define declares the collections of an owner type ("user", "page", "library" ...).
func Define(ownerType string, list ...Collection) {
	collectionsMutex.Lock()
	defer collectionsMutex.Unlock()
	if collections[ownerType] == nil {
		collections[ownerType] = map[string]Collection{}
	}
	for _, collection := range list {
		collections[ownerType][collection.Name] = collection
	}
}

// Find returns a declared collection.
func Find(ownerType, name string) (Collection, bool) {
	collectionsMutex.RLock()
	defer collectionsMutex.RUnlock()
	collection, ok := collections[ownerType][name]
	return collection, ok
}
