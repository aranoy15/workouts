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
var errInvalidMuscleGroupID = errors.New("Invalid muscle_group_id")
var errInvalidLevelID = errors.New("Invalid level_id")

type S3Service interface {
	UploadFile(ctx context.Context, objectID string, key string, body io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
	Bucket() string
}

type CreateExerciseRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	MuscleGroupID *string  `json:"muscle_group_id"`
	LevelID       *string  `json:"level_id"`
	VideoURLs     []string `json:"video_urls"`
}

type UpdateExerciseRequest struct {
	Name          *string   `json:"name"`
	Description   *string   `json:"description"`
	MuscleGroupID *string   `json:"muscle_group_id"`
	LevelID       *string   `json:"level_id"`
	VideoURLs     *[]string `json:"video_urls"`
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
	router.PUT("/exercises/:id", h.UpdateExercise)
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

	muscleGroupID, err := resolveMuscleGroupID(h.db, req.MuscleGroupID)
	if err != nil {
		writeFKError(c, err)
		return
	}
	levelID, err := resolveLevelID(h.db, req.LevelID)
	if err != nil {
		writeFKError(c, err)
		return
	}

	videoURLs := req.VideoURLs
	if videoURLs == nil {
		videoURLs = []string{}
	}

	exercise := &models.Exercise{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Description:   req.Description,
		MuscleGroupID: muscleGroupID,
		LevelID:       levelID,
		VideoURLs:     videoURLs,
	}

	if err := database.CreateExercise(h.db, exercise); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create exercise"))
		return
	}

	created, err := database.GetExerciseByID(h.db, exercise.ID)
	if err != nil {
		c.JSON(http.StatusCreated, models.NewSuccessResponse(exercise))
		return
	}
	c.JSON(http.StatusCreated, models.NewSuccessResponse(created))
}

func (h *ExerciseHandler) UpdateExercise(c *gin.Context) {
	exercise, err := database.GetExerciseByID(h.db, c.Param("id"))
	if err != nil {
		if errors.Is(err, database.ErrExerciseNotFound) {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to get exercise"))
		return
	}

	var req UpdateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body"))
		return
	}

	if req.Name != nil {
		if *req.Name == "" {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("Name is required"))
			return
		}
		exercise.Name = *req.Name
	}
	if req.Description != nil {
		exercise.Description = *req.Description
	}
	if req.MuscleGroupID != nil {
		muscleGroupID, err := resolveMuscleGroupID(h.db, req.MuscleGroupID)
		if err != nil {
			writeFKError(c, err)
			return
		}
		exercise.MuscleGroupID = muscleGroupID
	}
	if req.LevelID != nil {
		levelID, err := resolveLevelID(h.db, req.LevelID)
		if err != nil {
			writeFKError(c, err)
			return
		}
		exercise.LevelID = levelID
	}
	if req.VideoURLs != nil {
		exercise.VideoURLs = *req.VideoURLs
	}

	if err := database.UpdateExercise(h.db, exercise); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to update exercise"))
		return
	}

	updated, err := database.GetExerciseByID(h.db, exercise.ID)
	if err != nil {
		c.JSON(http.StatusOK, models.NewSuccessResponse(exercise))
		return
	}
	c.JSON(http.StatusOK, models.NewSuccessResponse(updated))
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

func resolveMuscleGroupID(db *gorm.DB, id *string) (*string, error) {
	if id == nil || strings.TrimSpace(*id) == "" {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*id)
	if _, err := database.GetMuscleGroupByID(db, trimmed); err != nil {
		if errors.Is(err, database.ErrMuscleGroupNotFound) {
			return nil, errInvalidMuscleGroupID
		}
		return nil, err
	}
	return &trimmed, nil
}

func resolveLevelID(db *gorm.DB, id *string) (*string, error) {
	if id == nil || strings.TrimSpace(*id) == "" {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*id)
	if _, err := database.GetLevelByID(db, trimmed); err != nil {
		if errors.Is(err, database.ErrLevelNotFound) {
			return nil, errInvalidLevelID
		}
		return nil, err
	}
	return &trimmed, nil
}

func writeFKError(c *gin.Context, err error) {
	if errors.Is(err, errInvalidMuscleGroupID) || errors.Is(err, errInvalidLevelID) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to validate references"))
}
