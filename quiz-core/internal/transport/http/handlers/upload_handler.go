package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxImageSize = 5 << 20 // 5 MB

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type UploadHandler struct {
	uploadsDir string
}

func NewUploadHandler(uploadsDir string) *UploadHandler {
	return &UploadHandler{uploadsDir: uploadsDir}
}

// UploadImage godoc
// @Summary      Загрузить изображение к вопросу
// @Tags         uploads
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Файл изображения (jpeg/png/gif/webp, до 5 МБ)"
// @Success      200  {object}  map[string]string  "url"
// @Failure      400  {object}  map[string]string  "error"
// @Failure      500  {object}  map[string]string  "error"
// @Security     BearerAuth
// @Router       /api/uploads/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxImageSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл слишком большой или не multipart"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}
	defer file.Close()

	if header.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл превышает 5 МБ"})
		return
	}

	// Определяем MIME по первым 512 байтам
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mime := http.DetectContentType(buf[:n])
	// Некоторые реализации возвращают параметры вроде "image/jpeg; charset=..."
	mime = strings.Split(mime, ";")[0]
	mime = strings.TrimSpace(mime)

	ext, ok := allowedMIME[mime]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неподдерживаемый тип файла: %s", mime)})
		return
	}

	if err := os.MkdirAll(h.uploadsDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать каталог загрузок"})
		return
	}

	filename := uuid.New().String() + ext
	dst := filepath.Join(h.uploadsDir, filename)

	// Перематываем файл в начало перед записью
	if seeker, ok := file.(interface {
		Seek(int64, int) (int64, error)
	}); ok {
		seeker.Seek(0, 0)
	}

	if err := c.SaveUploadedFile(header, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сохранения файла"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + filename})
}
