package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ayeshakhan-29/test-task-BE/internal/app/models"
	"github.com/ayeshakhan-29/test-task-BE/internal/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

type CrawlHandler struct {
	db *database.Database
}

func NewCrawlHandler(db *database.Database) *CrawlHandler {
	return &CrawlHandler{db: db}
}

func (h *CrawlHandler) CrawlURL(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
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

	// Check for debug mode
	debug := c.Query("debug") == "true"

	// Parse the URL to ensure it's valid
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
		URL:        req.URL,
		UserID:     userID.(uint64),
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

	// Count links
	internalLinks, externalLinks, inaccessibleLinks := countLinks(doc, parsedURL.Hostname())
	result.InternalLinks = internalLinks
	result.ExternalLinks = externalLinks
	result.InaccessibleLinks = inaccessibleLinks

	// Check for login form
	result.HasLoginForm = hasLoginForm(doc)

	// Save to database
	if err := h.db.Create(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save crawl result"})
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
func countLinks(doc *goquery.Document, hostname string) (int, int, int) {
	internal, external, inaccessible := 0, 0, 0

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" || strings.HasPrefix(href, "#") {
			return
		}

		// Check if link is internal or external
		u, err := url.Parse(href)
		if err != nil {
			return
		}

		if u.Hostname() == "" || u.Hostname() == hostname {
			internal++
		} else {
			external++
			// Check if external link is accessible
			if resp, err := http.Head(href); err == nil {
				if resp.StatusCode >= 400 {
					inaccessible++
				}
			}
		}
	})

	return internal, external, inaccessible
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
