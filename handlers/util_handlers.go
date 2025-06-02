package handlers

import (
	// Keep log for potential debug/error logging if needed
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vit0-9/utils_api/models"
	"github.com/vit0-9/utils_api/pkg/utils"
)

// URLUtilitiesHandlers groups URL specific utilities
type URLUtilitiesHandlers struct{}

func NewURLUtilitiesHandlers() *URLUtilitiesHandlers {
	return &URLUtilitiesHandlers{}
}

// CleanURLHandler (remains POST due to potentially long URL in body, but could be GET if URL is query param and not too long)
// For this refactor, let's keep it POST as per original but change the Swagger tag.
// If we were to change it to GET: c.Query("url")
// CleanURLHandler godoc
// @Summary      Clean a URL
// @Description  Removes known tracking parameters from a given URL.
// @Tags         URL Manipulation
// @Accept       json
// @Produce      json
// @Param        urlRequest body models.CleanURLRequest true "URL to clean"
// @Success      200 {object} models.DetailedCleanURLResponse
// @Failure      400 {object} map[string]string "Error: Invalid request payload"
// @Failure      500 {object} map[string]string "Error: Failed to process URL"
// @Router       /url/clean [post]
func (h *URLUtilitiesHandlers) CleanURLHandler(c *gin.Context) {
	var req models.CleanURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url field is required"})
		return
	}

	cleanResult, err := utils.CleanURL(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process URL for cleaning", "details": err.Error()})
		return
	}
	response := models.DetailedCleanURLResponse{
		OriginalURL:   models.SafeURLString(req.URL),
		CleanedURL:    models.SafeURLString(cleanResult.CleanedURL),
		RemovedParams: cleanResult.RemovedParams,
	}
	if len(cleanResult.RemovedParams) == 0 {
		response.Message = "No known tracking parameters found to remove."
	}
	c.JSON(http.StatusOK, response)
}

// ResolveRedirectHandler godoc
// @Summary      Resolve URL Redirects
// @Description  Follows HTTP redirects for a given URL (e.g., a shortlink) and returns the final destination URL.
// @Tags         URL Manipulation
// @Produce      json
// @Param        url query string true "URL to resolve"
// @Success      200 {object} models.ResolveRedirectResponse "Successfully resolved URL or error during resolution"
// @Failure      400 {object} map[string]string "Error: Invalid input (e.g., missing URL)"
// @Router       /url/resolve-redirect [get]
func (h *URLUtilitiesHandlers) ResolveRedirectHandler(c *gin.Context) {
	urlQuery := c.Query("url")
	if urlQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}

	finalURL, err := utils.ResolveRedirect(urlQuery) // Assuming utils.ResolveRedirect exists
	if err != nil {
		c.JSON(http.StatusOK, models.ResolveRedirectResponse{ // Still 200 but with error in body
			OriginalURL: models.SafeURLString(urlQuery),
			FinalURL:    models.SafeURLString(finalURL), // May be empty or last known on error
			Error:       err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, models.ResolveRedirectResponse{
		OriginalURL: models.SafeURLString(urlQuery),
		FinalURL:    models.SafeURLString(finalURL),
	})
}

// GenerateUTMHandler (remains POST due to complex input body)
// GenerateUTMHandler godoc
// @Summary      Generate UTM suffixed URLs
// @Description  Creates one or more URLs with UTM tracking parameters. Supports bulk creation and formatting options.
// @Tags         URL Manipulation
// @Accept       json
// @Produce      json
// @Param        utm_request body models.UTMGeneratorRequest true "UTM Generation Request"
// @Success      200 {object} models.UTMGeneratorResponse "Successfully generated UTM URLs"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Error during URL generation"
// @Router       /url/generate-utm [post]
func (h *URLUtilitiesHandlers) GenerateUTMHandler(c *gin.Context) {
	var req models.UTMGeneratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}
	if req.Options == nil {
		req.Options = &utils.UTMGeneratorOptions{}
	}

	var generatedLinks []models.GeneratedUTMLink
	for _, varSet := range req.VariableSets {
		fullParams := utils.FullUTMParams{
			Source:   varSet.Source,
			Medium:   varSet.Medium,
			Campaign: req.CommonParams.Campaign,
			Term:     req.CommonParams.Term,
			Content:  req.CommonParams.Content,
		}
		if varSet.Campaign != "" {
			fullParams.Campaign = varSet.Campaign
		}
		if varSet.Term != "" {
			fullParams.Term = varSet.Term
		}
		if varSet.Content != "" {
			fullParams.Content = varSet.Content
		}

		if fullParams.Source == "" || fullParams.Medium == "" || fullParams.Campaign == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "utm_source, utm_medium, and utm_campaign are required for each generated link."})
			return
		}

		finalURL, err := utils.GenerateUTMLink(req.BaseURL, fullParams, req.Options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate UTM link", "details": err.Error()})
			return
		}
		generatedLinks = append(generatedLinks, models.GeneratedUTMLink{
			Source:   utils.FormatUTMValue(fullParams.Source, req.Options),
			Medium:   utils.FormatUTMValue(fullParams.Medium, req.Options),
			Campaign: utils.FormatUTMValue(fullParams.Campaign, req.Options),
			Term:     utils.FormatUTMValue(fullParams.Term, req.Options),
			Content:  utils.FormatUTMValue(fullParams.Content, req.Options),
			FullURL:  models.SafeURLString(finalURL),
		})
	}
	response := models.UTMGeneratorResponse{
		BaseURL:        req.BaseURL, // This should be string, not SafeURLString
		GeneratedURLs:  generatedLinks,
		OptionsApplied: req.Options,
	}
	c.JSON(http.StatusOK, response)
}
