package handlers

import (
	"net/http"
	"strconv"

	"go-blog/models"

	"github.com/gin-gonic/gin"
)

// Index renders the public blog feed with optional tag filter and pagination.
// c is the Gin request context; query params page and tag drive listing.
func (h *Handler) Index(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit := 5
	offset := (page - 1) * limit

	tagFilter := c.Query("tag")

	var posts []models.Post
	query := h.DB.Preload("Tags").Where("hidden = ?", false).Order("created_at desc")

	if tagFilter != "" {
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.name = ?", tagFilter)
	}

	var total int64
	query.Model(&models.Post{}).Count(&total)

	query.Limit(limit).Offset(offset).Find(&posts)

	hasNext := int64(page*limit) < total

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	pages := make([]int, totalPages)
	for i := range pages {
		pages[i] = i + 1
	}

	h.HTML(c, http.StatusOK, "public/index.html", gin.H{
		"Posts":      posts,
		"Page":       page,
		"Pages":      pages,
		"TotalPages": totalPages,
		"HasNext":    hasNext,
		"Tag":        tagFilter,
	})
}

// ShowPost renders a single public post by ID when it exists and is not hidden.
// c is the Gin request context; the :id path parameter selects the post.
func (h *Handler) ShowPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var post models.Post
	if err := h.DB.Preload("Tags").Where("hidden = ?", false).First(&post, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	h.HTML(c, http.StatusOK, "public/post.html", gin.H{
		"Post": post,
	})
}
