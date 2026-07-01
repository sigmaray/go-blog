package postops

import (
	"fmt"
	"math/rand"
	"time"

	"go-blog/models"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, count int) (int, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	titles := []string{
		"Getting Started with Go",
		"Building REST APIs",
		"Database Migrations 101",
		"Understanding Goroutines",
		"Web Development Tips",
		"Docker for Developers",
		"Testing Best Practices",
		"Clean Code in Go",
		"Deploying to Production",
		"Introduction to Gin",
		"PostgreSQL Performance",
		"Session-Based Auth",
	}

	contents := []string{
		"This post covers the fundamentals and practical examples.",
		"A deep dive into patterns and common pitfalls.",
		"Step-by-step guide with code snippets.",
		"Everything you need to know to get productive quickly.",
		"Lessons learned from real-world projects.",
	}

	tagPool := []string{"go", "web", "tutorial", "news", "devops", "database", "api", "docker", "testing", "gin"}

	created := 0
	for i := 0; i < count; i++ {
		title := titles[rng.Intn(len(titles))] + fmt.Sprintf(" #%d", i+1)
		content := contents[rng.Intn(len(contents))]

		post := models.Post{
			Title:   title,
			Content: content,
		}

		tagCount := rng.Intn(3) + 1
		usedTags := make(map[string]bool)
		for j := 0; j < tagCount; j++ {
			tagName := tagPool[rng.Intn(len(tagPool))]
			if usedTags[tagName] {
				continue
			}
			usedTags[tagName] = true

			var tag models.Tag
			if err := db.Where("name = ?", tagName).FirstOrCreate(&tag, models.Tag{Name: tagName}).Error; err != nil {
				return created, fmt.Errorf("failed to create tag: %w", err)
			}
			post.Tags = append(post.Tags, tag)
		}

		if err := db.Create(&post).Error; err != nil {
			return created, fmt.Errorf("failed to create post: %w", err)
		}
		created++
	}

	return created, nil
}

func Clear(db *gorm.DB) (postsDeleted, tagsDeleted int64, err error) {
	if err := db.Exec("DELETE FROM post_tags").Error; err != nil {
		return 0, 0, fmt.Errorf("failed to clear post_tags: %w", err)
	}

	postResult := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Post{})
	if postResult.Error != nil {
		return 0, 0, fmt.Errorf("failed to clear posts: %w", postResult.Error)
	}

	tagResult := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Tag{})
	if tagResult.Error != nil {
		return postResult.RowsAffected, 0, fmt.Errorf("failed to clear tags: %w", tagResult.Error)
	}

	return postResult.RowsAffected, tagResult.RowsAffected, nil
}
