package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vit0-9/utils_api/models"
	"github.com/vit0-9/utils_api/pkg/utils"
)

// WebAnalysisHandlers groups web page analysis utilities
type WebAnalysisHandlers struct{}

func NewWebAnalysisHandlers() *WebAnalysisHandlers {
	return &WebAnalysisHandlers{}
}

// StackAnalyzerHandler godoc
// @Summary      Analyze technology stack of a website
// @Description  Fetches a URL and uses Wappalyzergo to identify technologies used.
// @Tags         Web Analysis
// @Produce      json
// @Param        url query string true "URL of the website to analyze"
// @Success      200 {object} models.StackAnalyzerResponse "Successfully analyzed stack or error during analysis"
// @Failure      400 {object} map[string]string "Error: Invalid input (e.g., missing URL)"
// @Router       /web/stack-analyzer [get]
func (h *WebAnalysisHandlers) StackAnalyzerHandler(c *gin.Context) {
	urlQuery := c.Query("url")
	if urlQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}

	utilTechInfo, finalURL, err := utils.AnalyzeStack(urlQuery)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "wappalyzer client not available") || strings.Contains(errMsg, "failed to initialize wappalyzer client") {
			c.JSON(http.StatusInternalServerError, models.StackAnalyzerResponse{
				RequestURL: urlQuery,
				Error:      "Technology stack analyzer is currently unavailable.",
			})
			log.Printf("StackAnalyzerHandler critical error: %v", err)
			return
		}
		c.JSON(http.StatusOK, models.StackAnalyzerResponse{ // Still 200 but with error in body
			RequestURL: urlQuery,
			FinalURL:   finalURL,
			Error:      errMsg,
		})
		return
	}

	responseTechnologies := make([]models.DetectedTechnology, len(utilTechInfo))
	for i, uti := range utilTechInfo {
		responseTechnologies[i] = models.DetectedTechnology{
			Name:        uti.Name,
			Version:     uti.Version,
			Categories:  uti.Categories,
			Description: uti.Description,
			Website:     uti.Website,
			Icon:        uti.Icon,
			CPE:         uti.CPE,
		}
	}
	response := models.StackAnalyzerResponse{
		RequestURL:   urlQuery,
		FinalURL:     finalURL,
		Technologies: responseTechnologies,
	}
	c.JSON(http.StatusOK, response)
}

// HTTPHeadersHandler godoc
// @Summary      View HTTP response headers for a URL
// @Description  Fetches and displays the HTTP response headers from a given URL. Uses the advanced HTTP client which follows redirects by default.
// @Tags         Web Analysis
// @Produce      json
// @Param        url query string true "URL to fetch headers from"
// @Param        method query string false "HTTP method (GET or HEAD). Note: FetchURL currently defaults to GET."
// @Success      200 {object} models.HTTPHeadersResponse "Successfully retrieved HTTP headers or error during fetch"
// @Failure      400 {object} map[string]string "Error: Invalid input (e.g., missing URL)"
// @Router       /web/http-headers [get]
func (h *WebAnalysisHandlers) HTTPHeadersHandler(c *gin.Context) {
	urlQuery := c.Query("url")
	if urlQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}

	// The `method` query parameter is noted in Swagger, but our current `utils.FetchURL`
	// defaults to GET. If you need to support different methods like HEAD specifically
	// for this handler, `utils.FetchURL` would need to be extended or a different
	// utility function called. For now, we'll assume GET via FetchURL.
	// String methodQuery := c.Query("method")

	fetchResult, err := utils.FetchURL(urlQuery) // Use the FetchURL utility

	if err != nil {
		// FetchURL returns a formatted error. We can pass it along.
		// It's good to return 200 OK for utility endpoints even if the underlying fetch failed,
		// with the error detailed in the JSON body.
		response := models.HTTPHeadersResponse{
			RequestURL: urlQuery,
			Error:      err.Error(),
		}
		if fetchResult != nil { // If fetchResult is not nil, some partial info might exist
			response.FinalURL = fetchResult.FinalURL
			response.StatusCode = fetchResult.StatusCode
			response.Status = fetchResult.Status
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Successfully fetched
	response := models.HTTPHeadersResponse{
		RequestURL: urlQuery,
		StatusCode: fetchResult.StatusCode,
		Status:     fetchResult.Status,
		Headers:    fetchResult.Headers,
		FinalURL:   fetchResult.FinalURL,
	}

	c.JSON(http.StatusOK, response)
}
