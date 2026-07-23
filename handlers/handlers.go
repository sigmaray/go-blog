package handlers

import (
	"net/http"
	"strconv"

	"go-blog/models"
	"go-blog/sanitize"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Handler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:       db,
		Validate: validator.New(),
	}
}

// --- Public Routes ---

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
	query := h.DB.Preload("Tags").Order("created_at desc")

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

	c.HTML(http.StatusOK, "public/index.html", gin.H{
		"Posts":      posts,
		"Page":       page,
		"Pages":      pages,
		"TotalPages": totalPages,
		"HasNext":    hasNext,
		"Tag":        tagFilter,
	})
}

func (h *Handler) ShowPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var post models.Post
	if err := h.DB.Preload("Tags").First(&post, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "public/post.html", gin.H{
		"Post": post,
	})
}

// --- Auth Routes ---

func (h *Handler) LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user") != nil {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}
	c.HTML(http.StatusOK, "admin/login.html", gin.H{})
}

func (h *Handler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := models.FindUserByUsername(h.DB, username)
	if err != nil {
		c.HTML(http.StatusOK, "admin/login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	if user == nil || !models.CheckPassword(user.PasswordHash, password) {
		c.HTML(http.StatusOK, "admin/login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user", user.Username)
	session.Save()
	c.Redirect(http.StatusFound, "/admin/")
}

func (h *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

// --- Admin Routes ---

func (h *Handler) AdminDashboard(c *gin.Context) {
	var posts []models.Post
	h.DB.Order("created_at desc").Find(&posts)

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"Posts": posts,
	})
}

func (h *Handler) NewPostPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/create_post.html", gin.H{})
}

type CreatePostInput struct {
	Title   string `form:"title"`
	Content string `form:"content" validate:"required"`
	Tags    string `form:"tags"`
}

func (h *Handler) CreatePost(c *gin.Context) {
	var input CreatePostInput
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "admin/create_post.html", gin.H{
			"Error":   "Invalid form data",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
		})
		return
	}

	if err := h.Validate.Struct(input); err != nil {
		c.HTML(http.StatusBadRequest, "admin/create_post.html", gin.H{
			"Error":   "Content is required",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
		})
		return
	}

	content := sanitize.HTML(input.Content)
	if content == "" {
		c.HTML(http.StatusBadRequest, "admin/create_post.html", gin.H{
			"Error":   "Content is required",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
		})
		return
	}

	post := models.Post{
		Title:   input.Title,
		Content: content,
	}

	tags, err := h.buildTags(input.Tags)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin/create_post.html", gin.H{
			"Error":   "Failed to process tags",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
		})
		return
	}
	post.Tags = tags

	if err := h.DB.Create(&post).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin/create_post.html", gin.H{
			"Error":   "Failed to create post",
			"Title":   input.Title,
			"Content": input.Content,
			"Tags":    input.Tags,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/")
}
