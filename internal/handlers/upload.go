package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/jmoiron/sqlx"
)

type UploadHandler struct {
	db             *sqlx.DB
	cloudinaryCloud   string
	cloudinaryKey     string
	cloudinarySecret  string
}

func NewUploadHandler(db *sqlx.DB, cloudinaryCloud, cloudinaryKey, cloudinarySecret string) *UploadHandler {
	return &UploadHandler{
		db:               db,
		cloudinaryCloud:  cloudinaryCloud,
		cloudinaryKey:    cloudinaryKey,
		cloudinarySecret: cloudinarySecret,
	}
}

func (h *UploadHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large (max 5MB)"})
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing avatar file"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only image files are allowed"})
		return
	}

	url, err := h.uploadToCloudinary(file, fmt.Sprintf("avatars/user_%d", claims.UserID))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to upload image"})
		return
	}

	_, err = h.db.Exec(`UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`, url, claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"avatar_url": url})
}

func (h *UploadHandler) uploadToCloudinary(file multipart.File, publicID string) (string, error) {
	if h.cloudinaryCloud == "" || h.cloudinaryKey == "" {
		return "", fmt.Errorf("cloudinary not configured")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("public_id", publicID)
	writer.WriteField("api_key", h.cloudinaryKey)
	writer.WriteField("timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	params := map[string]string{
		"public_id": publicID,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}
	signature := h.cloudinarySignature(params)
	writer.WriteField("signature", signature)

	part, _ := writer.CreateFormFile("file", "avatar.jpg")
	io.Copy(part, file)
	writer.Close()

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", h.cloudinaryCloud)
	resp, err := http.Post(uploadURL, writer.FormDataContentType(), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudinary upload failed: %s", string(respBody))
	}

	var result struct {
		SecureURL string `json:"secure_url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.SecureURL, nil
}

func (h *UploadHandler) cloudinarySignature(params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	parts = append(parts, h.cloudinarySecret)

	input := strings.Join(parts, "&")
	hsh := sha1.Sum([]byte(input))
	return hex.EncodeToString(hsh[:])
}
