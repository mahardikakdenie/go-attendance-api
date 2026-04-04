package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type MediaService interface {
	Upload(ctx context.Context, fileBytes []byte, filename string) (*model.Media, error)
}

type mediaService struct {
	repo repository.MediaRepository
}

func NewMediaService(repo repository.MediaRepository) MediaService {
	return &mediaService{repo: repo}
}

func forceDomain(url string) string {
	if url == "" {
		return url
	}
	return strings.ReplaceAll(url, "https://i.ibb.co", "https://i.ibb.co.com")
}

func (s *mediaService) Upload(ctx context.Context, fileBytes []byte, filename string) (*model.Media, error) {
	if len(fileBytes) == 0 {
		return nil, errors.New("file empty")
	}
	if os.Getenv("IMGBB_API_KEY") == "" {
		return nil, errors.New("IMGBB_API_KEY not configured")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return nil, err
	}

	if _, err := part.Write(fileBytes); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://api.imgbb.com/1/upload?key=%s", os.Getenv("IMGBB_API_KEY")),
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("imgbb upload failed: %s", string(respBody))
	}

	var result struct {
		Success bool       `json:"success"`
		Data    model.Meta `json:"data"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if !result.Success {
		if result.Error.Message != "" {
			return nil, errors.New(result.Error.Message)
		}
		return nil, errors.New("imgbb upload failed")
	}

	result.Data.URL = forceDomain(result.Data.URL)
	result.Data.Image.URL = forceDomain(result.Data.Image.URL)
	result.Data.Thumb.URL = forceDomain(result.Data.Thumb.URL)
	result.Data.DisplayURL = forceDomain(result.Data.DisplayURL)

	m := &model.Media{
		ImgbbID: &result.Data.ID,
		URL:     result.Data.URL,
		Type:    nil,
		Meta:    result.Data,
	}

	err = s.repo.Save(ctx, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
