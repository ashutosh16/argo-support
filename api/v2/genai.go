package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rs/zerolog/log"
	"github.intuit.com/dev-build/ibp-genai-service/internal/genai"
)

type GenAIService struct {
	analyzer genai.Analyzer
	feedback genai.Feedback
	activity genai.Activity
}

func NewGenAIService(analyzer genai.Analyzer, feedback genai.Feedback, activity genai.Activity) *GenAIService {
	return &GenAIService{
		analyzer: analyzer,
		feedback: feedback,
		activity: activity,
	}
}

func (s *GenAIService) RegisterHandlers(routes gin.IRoutes) {
	routes.POST("/analyze", s.AnalyzeFailures)
	routes.POST("/feedback", s.Feedback)
	routes.POST("/activity", s.Activity)
}

type Failure struct {
	Context string `json:"context"`
}

type AnalyzeFailuresRequest struct {
	Failures []*Failure `json:"failures"`
}

type FailureAnalysis struct {
	Analysis string `json:"analysis"`
}

type AnalyzeFailuresResponse struct {
	Analyses []*FailureAnalysis `json:"analyses"`
}

func (s *GenAIService) AnalyzeFailures(c *gin.Context) {
	// Parse request body
	var request AnalyzeFailuresRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleParseError(c, err)
		return
	}

	// The order in which we return the analyses must match the order of the failures in the request
	// We're using the following struct to keep track of the index of the initial failure
	type AnalysisResult struct {
		Index    int
		Analysis *FailureAnalysis
	}

	type ErrorResult struct {
		Index int
		Error error
	}

	// Channels for receiving analyses and errors with an index
	analysisChan := make(chan AnalysisResult)
	errChan := make(chan ErrorResult)

	// Tie the context to the request context
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Pre-allocation of the result slice
	analyses := make([]*FailureAnalysis, len(request.Failures))

	// Launch a goroutine for each failure analysis
	for i, failure := range request.Failures {
		go func(index int, failure *Failure) {
			analysis, err := s.analyzer.Analyze(ctx, failure.Context)
			if err != nil {
				select {
				case errChan <- ErrorResult{Index: index, Error: fmt.Errorf("Failed to analyze failure: %v", failure)}:
				case <-ctx.Done(): // Ensure goroutine can exit if context is cancelled
				}
				return
			}
			select {
			case analysisChan <- AnalysisResult{Index: index, Analysis: &FailureAnalysis{Analysis: analysis}}:
			case <-ctx.Done(): // Ensure goroutine can exit if context is cancelled
			}
		}(i, failure)
	}

	// Collect analyses and handle errors
	for i := 0; i < len(request.Failures); i++ {
		select {
		case analysis := <-analysisChan:
			analyses[analysis.Index] = analysis.Analysis
		case err := <-errChan:
			analyses[err.Index] = &FailureAnalysis{Analysis: fmt.Sprintf("Failed to analyze failure: %v", err.Error)}
			log.Error().Err(err.Error).Msg("Failed to analyze failure")
		case <-c.Request.Context().Done():
			log.Error().Msg("Analysis cancelled or timed out")
		}
	}

	// Build response
	response := AnalyzeFailuresResponse{Analyses: analyses}

	// Return the response as JSON
	log.Info().Msgf("Successfully analyzed failures")
	c.JSON(http.StatusOK, response)
}

type UserFeedbackRequest struct {
	FeedbackFor string `json:"feedbackFor"`
	Positive    string `json:"positive"`
	Comments    string `json:"comments"`
	BuildUrl    string `json:"buildUrl"`
	UserId      string `json:"userId"`
}

type UserFeedbackRequestFailure struct {
	UserFeedbackRequest
	FailureContext       string `json:"failureContext"`
	FailureAnalysis      string `json:"failureAnalysis"`
	FailureResolution    string `json:"failureResolution"`
	PipelineStepsNodeUrl string `json:"pipelineStepsNodeUrl"`
}

const (
	FeedbackForFailure = "failure"
)

func (s *GenAIService) Feedback(c *gin.Context) {
	// Parse request body
	var request UserFeedbackRequest
	var err error

	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		handleParseError(c, err)
		return
	}

	var feedbackData interface{}

	// Depending on the type of feedback, marshal into the appropriate struct to make sure the request is valid
	switch request.FeedbackFor {
	case FeedbackForFailure:
		// check if the request can be parsed into a UserFeedbackRequestFailure
		var failureRequest UserFeedbackRequestFailure
		err = c.ShouldBindBodyWith(&failureRequest, binding.JSON)
		feedbackData = failureRequest
	default:
		// this type of feedback isn't recognized
		err = fmt.Errorf("unexpected feedbackFor: %s", request.FeedbackFor)
	}

	// If we were unable to parse the feedback, return an error
	if err != nil {
		handleParseError(c, err)
		return
	}

	// Store the feedback
	err = s.feedback.Store(feedbackData)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store user feedback")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to store user feedback",
			"error":   err.Error(),
		})
		return
	}

	log.Info().Msg("Successfully stored user feedback")
	c.Status(http.StatusOK)
}

func (s *GenAIService) Activity(c *gin.Context) {
	// Parse request body
	jsonData := make(map[string]string)

	// Bind JSON to the map
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		handleParseError(c, err)
		return
	}

	// Store the feedback
	if err := s.activity.Store(jsonData); err != nil {
		log.Error().Err(err).Msg("Failed to store activity")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to store activity",
			"error":   err.Error(),
		})
		return
	}

	log.Info().Msg("Successfully stored activity")
	c.Status(http.StatusOK)
}

func handleParseError(c *gin.Context, err error) {
	log.Error().Err(err).Msg("Failed to parse request body")
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"message": "Failed to parse request body",
		"error":   err.Error(),
	})
}
