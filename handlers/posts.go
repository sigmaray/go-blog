package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"go-blog/models"
	"go-blog/sanitize"

	"github.com/gin-gonic/gin"
)

func (h *Handler) EditPostPage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}

	var post models.Post
	if err := h.DB.Preload("Tags").First(&post, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}

	c.HTML(http.StatusOK, "admin/edit_post.html", gin.H{
		"Post":   post,
		"Tags":   formatTagNames(post.Tags),
		"Hidden": post.Hidden,
	})
}

// UpdatePostInput holds validated form fields for updating a post.
// Hidden is true when the admin checkbox "Hide post" is checked.
type UpdatePostInput struct {
	Title   string `form:"title"`
	Content string `form:"content" validate:"required"`
	Tags    string `form:"tags"`
	Hidden  bool   `form:"hidden"`
}

func (h *Handler) UpdatePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}

	var post models.Post
	if err := h.DB.Preload("Tags").First(&post, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}

	var input UpdatePostInput
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "admin/edit_post.html", gin.H{
			"Error":   "Invalid form data",
			"Post":    post,
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	if err := h.Validate.Struct(input); err != nil {
		c.HTML(http.StatusBadRequest, "admin/edit_post.html", gin.H{
			"Error":   "Content is required",
			"Post":    post,
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	content := sanitize.HTML(input.Content)
	if content == "" {
		c.HTML(http.StatusBadRequest, "admin/edit_post.html", gin.H{
			"Error":   "Content is required",
			"Post":    post,
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	post.Title = input.Title
	post.Content = content
	post.Hidden = input.Hidden

	tags, err := h.buildTags(input.Tags)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin/edit_post.html", gin.H{
			"Error":   "Failed to process tags",
			"Post":    post,
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	if err := h.DB.Save(&post).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin/edit_post.html", gin.H{
			"Error":   "Failed to update post",
			"Post":    post,
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	if err := h.DB.Model(&post).Association("Tags").Replace(tags); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/edit_post.html", gin.H{
			"Error":   "Failed to update tags",
			"Post":    post,
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/")
}

func (h *Handler) DeletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}

	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}

	h.DB.Select("Tags").Delete(&post)
	c.Redirect(http.StatusFound, "/admin/")
}

func (h *Handler) buildTags(tagsInput string) ([]models.Tag, error) {
	if tagsInput == "" {
		return nil, nil
	}

	tagNames := strings.Split(tagsInput, ",")
	var tags []models.Tag

	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var tag models.Tag
		if err := h.DB.Where("name = ?", name).FirstOrCreate(&tag, models.Tag{Name: name}).Error; err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func formatTagNames(tags []models.Tag) string {
	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Name
	}
	return strings.Join(names, ", ")
}
