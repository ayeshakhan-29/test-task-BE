package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ayeshakhan-29/test-task-BE/internal/app/models"
	"github.com/ayeshakhan-29/test-task-BE/internal/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

func (h *CrawlHandler) ListCrawls(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var crawls []models.CrawlResult
	if err := h.db.DB.Where("user_id = ?", userID).Find(&crawls).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch crawl results"})
		return
	}

	response := make([]models.CrawlListResponse, 0, len(crawls))
	for _, crawl := range crawls {
		// Convert the JSON value to a slice if it's a number
		if len(crawl.InaccessibleLinks) == 0 {
			crawl.InaccessibleLinks = make(models.StringSlice, 0)
		}
		response = append(response, models.CrawlListResponse{
			ID:              crawl.ID,
			URL:             crawl.URL,
			PageTitle:       crawl.PageTitle,
			CreatedAt:       crawl.CreatedAt,
			HTMLVersion:     crawl.HTMLVersion,
			Headings:        crawl.Headings,
			InternalLinks:   crawl.InternalLinks,
			ExternalLinks:   crawl.ExternalLinks,
			InaccessibleLinks: crawl.InaccessibleLinks,
			HasLoginForm:    crawl.HasLoginForm,
		})
	}
	c.JSON(http.StatusOK, response)
}

func (h *CrawlHandler) GetCrawlByID(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get crawl ID from URL parameter
	crawlID := c.Param("id")
	if crawlID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crawl ID is required"})
		return
	}

	// Convert string ID to uint64
	crawlIDUint, err := strconv.ParseUint(crawlID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid crawl ID format"})
		return
	}

	// Find the crawl record
	var crawl models.CrawlResult
	if err := h.db.DB.First(&crawl, "id = ?", crawlIDUint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Crawl not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	// Check if user owns this crawl
	if crawl.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this crawl"})
		return
	}

	// Convert to response format
	response := models.CrawlListResponse{
		ID:              crawl.ID,
		URL:             crawl.URL,
		PageTitle:       crawl.PageTitle,
		CreatedAt:       crawl.CreatedAt,
		HTMLVersion:     crawl.HTMLVersion,
		Headings:        crawl.Headings,
		InternalLinks:   crawl.InternalLinks,
		ExternalLinks:   crawl.ExternalLinks,
		InaccessibleLinks: crawl.InaccessibleLinks,
		HasLoginForm:    crawl.HasLoginForm,
	}

	c.JSON(http.StatusOK, response)
}

func (h *CrawlHandler) DeleteCrawl(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get crawl ID from URL parameter
	crawlID := c.Param("id")
	if crawlID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crawl ID is required"})
		return
	}

	// Find the crawl record
	var crawl models.CrawlResult
	if err := h.db.DB.First(&crawl, crawlID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Crawl not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch crawl"})
		return
	}

	// Check if user owns this crawl
	if crawl.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this crawl"})
		return
	}

	// Delete the crawl
	if err := h.db.DB.Delete(&crawl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete crawl"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Crawl deleted successfully"})
}

type CrawlHandler struct {
	db *database.Database
}

func NewCrawlHandler(db *database.Database) *CrawlHandler {
	return &CrawlHandler{db: db}
}

func (h *CrawlHandler) BulkDeleteCrawls(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Validate that we have IDs to delete
	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided in request"})
		return
	}

	// Start a transaction
	db := h.db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
		}
	}()

	// Verify ownership of all records
	var invalidIDs []uint
	for _, id := range req.IDs {
		var count int64
		if err := db.Model(&models.CrawlResult{}).
			Where("id = ? AND user_id = ?", id, userID).
			Count(&count).
			Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
			return
		}

		if count == 0 {
			invalidIDs = append(invalidIDs, id)
		}
	}

	if len(invalidIDs) > 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to delete some crawls",
			"invalid_ids": invalidIDs,
		})
		return
	}

	// Delete the records
	result := db.Where("id IN (?) AND user_id = ?", req.IDs, userID).Delete(&models.CrawlResult{})
	if result.Error != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete crawls"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No crawls found to delete"})
		return
	}

	if err := db.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Crawls deleted successfully",
		"deleted_count": result.RowsAffected,
	})
}

func (h *CrawlHandler) CrawlURL(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	debug := c.Query("debug") == "true"

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	// Fetch the page
	resp, err := http.Get(parsedURL.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch URL: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read body for debug mode
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	// Create a fresh reader for goquery
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse HTML"})
		return
	}

	// Extract data
	result := models.CrawlResult{
		URL:    req.URL,
		UserID: userID.(uint64),
		HTMLVersion: func() string {
			version, err := extractHTMLVersionFromURL(req.URL)
			if err != nil {
				return ""
			}
			return version
		}(),
		PageTitle:   doc.Find("title").Text(),
	}

	// Count headings
	headings := models.HeadingCounts{
		H1: doc.Find("h1").Length(),
		H2: doc.Find("h2").Length(),
		H3: doc.Find("h3").Length(),
		H4: doc.Find("h4").Length(),
		H5: doc.Find("h5").Length(),
		H6: doc.Find("h6").Length(),
	}
	result.Headings = headings

	// Count links and get broken links
	internalLinks, externalLinks, _, brokenLinks := countLinks(doc, parsedURL.Hostname())
	result.InternalLinks = internalLinks
	result.ExternalLinks = externalLinks
	result.InaccessibleLinks = brokenLinks

	// Check for login form
	result.HasLoginForm = hasLoginForm(doc)

	// Check if a crawl entry with the same URL and user ID already exists
	var existingCrawl models.CrawlResult
	err = h.db.DB.Where("url = ? AND user_id = ?", req.URL, userID).First(&existingCrawl).Error

	if err == nil {
		// Update existing crawl
		result.ID = existingCrawl.ID
		result.CreatedAt = existingCrawl.CreatedAt // Preserve original creation time
		if err := h.db.DB.Save(&result).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update crawl result"})
			return
		}
	} else if err == gorm.ErrRecordNotFound {
		// Create new crawl if not exists
		if err := h.db.DB.Create(&result).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save crawl result"})
			return
		}
	} else {
		// Handle other database errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	// Return result with HTML if in debug mode
	if debug {
		c.JSON(http.StatusOK, gin.H{
			"result": result,
			"html":   string(body),
		})
	} else {
		c.JSON(http.StatusOK, result)
	}
}

func extractHTMLVersionFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return "Unknown", nil // No doctype found
		case html.DoctypeToken:
			doctype := strings.ToLower(string(z.Text()))
			switch {
			case doctype == "html":
				return "HTML5", nil
			case strings.Contains(doctype, "xhtml 1.0 strict"):
				return "XHTML 1.0 Strict", nil
			case strings.Contains(doctype, "xhtml 1.0 transitional"):
				return "XHTML 1.0 Transitional", nil
			case strings.Contains(doctype, "xhtml 1.0 frameset"):
				return "XHTML 1.0 Frameset", nil
			case strings.Contains(doctype, "xhtml 1.1"):
				return "XHTML 1.1", nil
			case strings.Contains(doctype, "html 4.01 transitional"):
				return "HTML 4.01 Transitional", nil
			case strings.Contains(doctype, "html 4.01 frameset"):
				return "HTML 4.01 Frameset", nil
			case strings.Contains(doctype, "html 4.01"):
				return "HTML 4.01", nil
			case strings.Contains(doctype, "html 4.0"):
				return "HTML 4.0", nil
			case strings.Contains(doctype, "html 3.2"):
				return "HTML 3.2", nil
			case strings.Contains(doctype, "html 2.0"):
				return "HTML 2.0", nil
			default:
				return "Unknown", nil
			}
		}
	}
}

func countLinks(doc *goquery.Document, hostname string) (int, int, int, []string) {
	internal, external, inaccessible := 0, 0, 0
	var brokenLinks []string

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" || strings.HasPrefix(href, "#") {
			return
		}

		// Check if link is accessible
		resp, err := http.Head(href)
		if err != nil || resp.StatusCode >= 400 {
			brokenLinks = append(brokenLinks, href)
			inaccessible++
			return
		}

		if strings.HasPrefix(href, "http") {
			if strings.Contains(href, hostname) {
				internal++
			} else {
				external++
			}
		} else {
			internal++
		}
	})

	return internal, external, inaccessible, brokenLinks
}

func hasLoginForm(doc *goquery.Document) bool {
	// Check for common login form indicators
	loginSelectors := []string{
		"input[type='password']",
		"form[action*='login']",
		"form[action*='signin']",
		"#login-form",
		".login-form",
	}

	for _, selector := range loginSelectors {
		if doc.Find(selector).Length() > 0 {
			return true
		}
	}

	return false
}
