package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"workouts-backend/src/database"
	"workouts-backend/src/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	maxVideoSize   = 100 << 20 // 100 MB
	videoFormField = "video"
	videosPrefix   = "videos"
)

var errInvalidVideoURL = errors.New("invalid video_url")

type S3Service interface {
	UploadFile(ctx context.Context, objectID string, key string, body io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
	Bucket() string
}

type CreateExerciseRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	MuscleGroup string `json:"muscle_group"`
	Level       string `json:"level"`
	VideoURL    string `json:"video_url"`
}

type DeleteVideoRequest struct {
	Key      string `json:"key"`
	VideoURL string `json:"video_url"`
}

type ExerciseHandler struct {
	db *gorm.DB
	s3 S3Service
}

func NewExerciseHandler(db *gorm.DB, s3Client S3Service) *ExerciseHandler {
	return &ExerciseHandler{db: db, s3: s3Client}
}

func RegisterExerciseHandler(router *gin.RouterGroup, db *gorm.DB) {
	h := NewExerciseHandler(db, nil)
	router.GET("/exercises", h.GetExercises)
	router.GET("/exercises/:id", h.GetExercise)
}

func RegisterAdminExerciseHandler(router *gin.RouterGroup, db *gorm.DB, s3Client S3Service) {
	h := NewExerciseHandler(db, s3Client)
	router.POST("/exercises", h.CreateExercise)
	router.DELETE("/exercises/:id", h.DeleteExercise)
	router.POST("/videos", h.UploadVideo)
	router.DELETE("/videos", h.DeleteVideo)
}

func (h *ExerciseHandler) GetExercises(c *gin.Context) {
	exercises, err := database.GetExercises(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get exercises"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(exercises))
}

func (h *ExerciseHandler) GetExercise(c *gin.Context) {
	exercise, err := database.GetExerciseByID(h.db, c.Param("id"))
	if err != nil {
		if errors.Is(err, database.ErrExerciseNotFound) {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get exercise"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(exercise))
}

func (h *ExerciseHandler) CreateExercise(c *gin.Context) {
	var req CreateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
		return
	}

	exercise := &models.Exercise{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		MuscleGroup: req.MuscleGroup,
		Level:       req.Level,
		VideoURL:    req.VideoURL,
	}

	if err := database.CreateExercise(h.db, exercise); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create exercise"))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(exercise))
}

func (h *ExerciseHandler) DeleteExercise(c *gin.Context) {
	if err := database.DeleteExercise(h.db, c.Param("id")); err != nil {
		if errors.Is(err, database.ErrExerciseNotFound) {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to delete exercise"))
		return
	}

	c.JSON(http.StatusOK, models.NewMessageResponse("Exercise deleted"))
}

func (h *ExerciseHandler) UploadVideo(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxVideoSize); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Failed to parse multipart form"))
		return
	}

	file, header, err := c.Request.FormFile(videoFormField)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("video file is required"))
		return
	}
	defer file.Close()

	filename := path.Base(header.Filename)
	if filename == "." || filename == "/" || filename == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid filename"))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectID := videosPrefix + "/" + uuid.New().String()
	videoURL, err := h.s3.UploadFile(c.Request.Context(), objectID, filename, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to upload video to S3"))
		return
	}

	key := objectID + "/" + filename
	c.JSON(http.StatusCreated, models.NewSuccessResponse(gin.H{
		"key":       key,
		"video_url": videoURL,
	}))
}

func (h *ExerciseHandler) DeleteVideo(c *gin.Context) {
	key := c.Query("key")
	videoURL := c.Query("video_url")

	if strings.Contains(c.GetHeader("Content-Type"), "application/json") && c.Request.ContentLength != 0 {
		var req DeleteVideoRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
			return
		}
		if req.Key != "" {
			key = req.Key
		}
		if req.VideoURL != "" {
			videoURL = req.VideoURL
		}
	}

	if key == "" && videoURL != "" {
		parsed, err := keyFromVideoURL(videoURL, h.s3.Bucket())
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid video_url"))
			return
		}
		key = parsed
	}

	if key == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("key or video_url is required"))
		return
	}

	key = strings.TrimLeft(key, "/")
	if !strings.HasPrefix(key, videosPrefix+"/") {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("invalid object key"))
		return
	}

	if err := h.s3.DeleteFile(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to delete video from S3"))
		return
	}

	c.JSON(http.StatusOK, models.NewMessageResponse("Video deleted"))
}

func keyFromVideoURL(rawURL, bucket string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	p := strings.Trim(u.Path, "/")
	parts := strings.SplitN(p, "/", 2)
	if len(parts) != 2 || parts[0] != bucket || parts[1] == "" {
		return "", errInvalidVideoURL
	}
	return parts[1], nil
}
