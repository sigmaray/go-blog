package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"go-blog/postops"

	"github.com/gin-gonic/gin"
)

type SQLInput struct {
	Query string `form:"query"`
}

func (h *Handler) ToolsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/tools.html", gin.H{})
}

func (h *Handler) ExecuteSQL(c *gin.Context) {
	var input SQLInput
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
			"Error": "Invalid form data",
		})
		return
	}

	query := strings.TrimSpace(input.Query)
	if query == "" {
		c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
			"Error": "Query is required",
			"Query": input.Query,
		})
		return
	}

	if hasMultipleStatements(query) {
		c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
			"Error": "Only one SQL statement is allowed",
			"Query": input.Query,
		})
		return
	}

	data := gin.H{"Query": input.Query}

	if isReadQuery(query) {
		columns, rows, err := h.scanQueryResults(query)
		if err != nil {
			c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
				"Error": err.Error(),
				"Query": input.Query,
			})
			return
		}
		data["Columns"] = columns
		data["Rows"] = rows
		data["RowCount"] = len(rows)
	} else {
		result := h.DB.Exec(query)
		if result.Error != nil {
			c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
				"Error": result.Error.Error(),
				"Query": input.Query,
			})
			return
		}
		data["IsMutation"] = true
		data["RowsAffected"] = result.RowsAffected
	}

	c.HTML(http.StatusOK, "admin/tools.html", data)
}

type PostsSeedInput struct {
	Count int `form:"count"`
}

func (h *Handler) PostsSeed(c *gin.Context) {
	var input PostsSeedInput
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
			"Error": "Invalid form data",
		})
		return
	}

	count := input.Count
	if count < 1 {
		count = 10
	}

	created, err := postops.Seed(h.DB, count)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "admin/tools.html", gin.H{
		"Message": fmt.Sprintf("Created %d post(s).", created),
	})
}

func (h *Handler) PostsClear(c *gin.Context) {
	postsDeleted, tagsDeleted, err := postops.Clear(h.DB)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/tools.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "admin/tools.html", gin.H{
		"Message": fmt.Sprintf("Deleted %d post(s) and %d tag(s).", postsDeleted, tagsDeleted),
	})
}

func isReadQuery(sql string) bool {
	fields := strings.Fields(strings.TrimSpace(sql))
	if len(fields) == 0 {
		return false
	}
	switch strings.ToUpper(fields[0]) {
	case "SELECT", "WITH", "SHOW", "EXPLAIN", "TABLE", "VALUES":
		return true
	default:
		return false
	}
}

func hasMultipleStatements(sql string) bool {
	parts := strings.Split(strings.TrimSpace(sql), ";")
	nonEmpty := 0
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			nonEmpty++
		}
	}
	return nonEmpty > 1
}

func (h *Handler) scanQueryResults(sql string) ([]string, [][]string, error) {
	rows, err := h.DB.Raw(sql).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var result [][]string
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		row := make([]string, len(columns))
		for i, v := range values {
			row[i] = formatSQLValue(v)
		}
		result = append(result, row)
	}

	return columns, result, rows.Err()
}

func formatSQLValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return fmt.Sprint(val)
	}
}
