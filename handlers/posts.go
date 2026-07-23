package handlers

import (
	"net/http"
	"strconv"

	"go-blog/models"
	"go-blog/sanitize"

	"github.com/gin-gonic/gin"
)

// AdminDashboard lists all posts for the authenticated admin.
// c is the Gin request context used to render the dashboard.
func (h *Handler) AdminDashboard(c *gin.Context) {
	var posts []models.Post
	h.DB.Order("created_at desc").Find(&posts)

	h.HTML(c, http.StatusOK, "admin/dashboard.html", gin.H{
		"Posts": posts,
	})
}

// NewPostPage renders the create-post form.
// c is the Gin request context used to render the page.
func (h *Handler) NewPostPage(c *gin.Context) {
	h.HTML(c, http.StatusOK, "admin/create_post.html", gin.H{})
}

// CreatePostInput holds validated form fields for creating a post.
// Hidden is true when the admin checkbox "Hide post" is checked.
type CreatePostInput struct {
	Title   string `form:"title"`
	Content string `form:"content" validate:"required"`
	Tags    string `form:"tags"`
	Hidden  bool   `form:"hidden"`
}

// CreatePost validates input, sanitizes HTML, resolves tags, and inserts a post.
// c is the Gin request context carrying the create-post form POST body.
func (h *Handler) CreatePost(c *gin.Context) {
	var input CreatePostInput
	if err := c.ShouldBind(&input); err != nil {
		h.HTML(c, http.StatusBadRequest, "admin/create_post.html", gin.H{
			"Error":   "Invalid form data",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	if err := h.Validate.Struct(input); err != nil {
		h.HTML(c, http.StatusBadRequest, "admin/create_post.html", gin.H{
			"Error":   "Content is required",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	content := sanitize.HTML(input.Content)
	if content == "" {
		h.HTML(c, http.StatusBadRequest, "admin/create_post.html", gin.H{
			"Error":   "Content is required",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	post := models.Post{
		Title:   input.Title,
		Content: content,
		Hidden:  input.Hidden,
	}

	tags, err := models.FindOrCreateTags(h.DB, input.Tags)
	if err != nil {
		h.HTML(c, http.StatusInternalServerError, "admin/create_post.html", gin.H{
			"Error":   "Failed to process tags",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}
	post.Tags = tags

	if err := h.DB.Create(&post).Error; err != nil {
		h.HTML(c, http.StatusInternalServerError, "admin/create_post.html", gin.H{
			"Error":   "Failed to create post",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
			"Hidden":  input.Hidden,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/")
}

// EditPostPage loads a post and renders the edit form.
// c is the Gin request context; the :id path parameter selects the post.
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

	h.HTML(c, http.StatusOK, "admin/edit_post.html", gin.H{
		"Post":   post,
		"Tags":   models.FormatTagNames(post.Tags),
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

// UpdatePost validates input, sanitizes HTML, updates the post, and replaces tags.
// c is the Gin request context; :id selects the post and the body carries form fields.
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
		h.HTML(c, http.StatusBadRequest, "admin/edit_post.html", gin.H{
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
		h.HTML(c, http.StatusBadRequest, "admin/edit_post.html", gin.H{
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
		h.HTML(c, http.StatusBadRequest, "admin/edit_post.html", gin.H{
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

	tags, err := models.FindOrCreateTags(h.DB, input.Tags)
	if err != nil {
		h.HTML(c, http.StatusInternalServerError, "admin/edit_post.html", gin.H{
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
		h.HTML(c, http.StatusInternalServerError, "admin/edit_post.html", gin.H{
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
		h.HTML(c, http.StatusInternalServerError, "admin/edit_post.html", gin.H{
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

// DeletePost removes a post (and its tag associations) then redirects to admin.
// c is the Gin request context; the :id path parameter selects the post.
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
