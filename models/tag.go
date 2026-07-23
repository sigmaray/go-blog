package models

import (
	"strings"

	"gorm.io/gorm"
)

// FindOrCreateTags parses a comma-separated tag list and returns Tag records,
// creating any names that do not already exist in the database.
// db is the GORM handle used for lookups and inserts.
// tagsInput is the raw form value (for example "go, web, tutorial").
func FindOrCreateTags(db *gorm.DB, tagsInput string) ([]Tag, error) {
	if tagsInput == "" {
		return nil, nil
	}

	tagNames := strings.Split(tagsInput, ",")
	var tags []Tag

	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var tag Tag
		if err := db.Where("name = ?", name).FirstOrCreate(&tag, Tag{Name: name}).Error; err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// FormatTagNames joins tag names into a comma-separated string for edit forms.
// tags is the slice of Tag models whose names should be displayed.
func FormatTagNames(tags []Tag) string {
	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Name
	}
	return strings.Join(names, ", ")
}
