package journal

import "path/filepath"

func GetJournalFilename(filenameTextIn string) string {
	dir := filepath.Dir(filenameTextIn)
	name := "." + filepath.Base(filenameTextIn) + ".journal"
	journalPath := filepath.Join(dir, name)
	return journalPath
}
