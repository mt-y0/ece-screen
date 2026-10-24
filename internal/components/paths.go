package components

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
)

// storagePath returns the on-disk path of an uploaded file, rooted at
// PocketBase's data dir rather than a hard-coded "pb_data" relative to the CWD.
// A missing file is logged (the caller still gets the path).
func storagePath(pb *pocketbase.PocketBase, collectionID, recordID, file string) string {
	p := filepath.Join(pb.DataDir(), "storage", collectionID, recordID, file)
	if _, err := os.Stat(p); err != nil {
		log.Printf("media file missing: %s", p)
	}
	return p
}
