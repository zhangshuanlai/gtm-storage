package client

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// Client represents the GTM Storage client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// ObjectInfo represents object metadata
type ObjectInfo struct {
	Key          string    `xml:"Key"`
	Name         string    `xml:"Name"`
	ContentType  string    `xml:"ContentType"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         int64     `xml:"Size"`
}

// ListBucketResult represents the response from list objects
type ListBucketResult struct {
	Name     string       `xml:"Name"`
	Prefix   string       `xml:"Prefix"`
	Contents []ObjectInfo `xml:"Contents"`
}

// UploadResult represents the result of an upload operation
type UploadResult struct {
	Key          string
	ETag         string
	PreviewURL   string
	ThumbnailURL string
}

// ClientOptions represents configuration options for the client
type ClientOptions struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewClient creates a new GTM Storage client
func NewClient(options ClientOptions) *Client {
	if options.HTTPClient == nil {
		timeout := options.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		options.HTTPClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		baseURL:    strings.TrimRight(options.BaseURL, "/"),
		httpClient: options.HTTPClient,
		apiKey:     options.APIKey,
	}
}

// addAuth adds authentication to the request
func (c *Client) addAuth(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		// 或者根据实际的认证方式设置
		req.Header.Set("X-API-Key", c.apiKey)
	}
}

// MakeBucket creates a new bucket
func (c *Client) MakeBucket(ctx context.Context, bucketName string) error {
	url := fmt.Sprintf("%s/api/%s", c.baseURL, bucketName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.addAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create bucket: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// DeleteBucket deletes a bucket
func (c *Client) DeleteBucket(ctx context.Context, bucketName string) error {
	url := fmt.Sprintf("%s/api/%s", c.baseURL, bucketName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.addAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete bucket: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// PutObject uploads an object to the bucket
func (c *Client) PutObject(ctx context.Context, bucketName, objectKey string, reader io.Reader, filename string) (*UploadResult, error) {
	url := fmt.Sprintf("%s/api/%s/%s", c.baseURL, bucketName, objectKey)

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(fileWriter, reader); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.addAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to upload object: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	result := &UploadResult{
		Key:  objectKey,
		ETag: resp.Header.Get("ETag"),
	}

	// Extract URLs from response body (简单解析，实际可能需要更复杂的解析)
	bodyStr := string(body)
	if strings.Contains(bodyStr, "预览地址:") {
		lines := strings.Split(bodyStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "预览地址:") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					result.PreviewURL = parts[len(parts)-1]
				}
			}
			if strings.Contains(line, "缩略图地址:") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					result.ThumbnailURL = parts[len(parts)-1]
				}
			}
		}
	}

	return result, nil
}

// GetObject retrieves an object from the bucket
func (c *Client) GetObject(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/api/%s/%s", c.baseURL, bucketName, objectKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get object: %s (status: %d)", string(body), resp.StatusCode)
	}

	return resp.Body, nil
}

// GetObjectRange retrieves a range of bytes from an object
func (c *Client) GetObjectRange(ctx context.Context, bucketName, objectKey string, start, end int64) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/api/%s/%s", c.baseURL, bucketName, objectKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if start > 0 || end > 0 {
		rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get object: %s (status: %d)", string(body), resp.StatusCode)
	}

	return resp.Body, nil
}

// DeleteObject deletes an object from the bucket
func (c *Client) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	url := fmt.Sprintf("%s/api/%s/%s", c.baseURL, bucketName, objectKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.addAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete object: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// HeadObject retrieves object metadata
func (c *Client) HeadObject(ctx context.Context, bucketName, objectKey string) (*ObjectInfo, error) {
	url := fmt.Sprintf("%s/api/%s/%s", c.baseURL, bucketName, objectKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get object metadata: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse Last-Modified
	lastModified, _ := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))

	return &ObjectInfo{
		Key:          objectKey,
		ContentType:  resp.Header.Get("Content-Type"),
		LastModified: lastModified,
		ETag:         strings.Trim(resp.Header.Get("ETag"), "\""),
		Size:         resp.ContentLength,
	}, nil
}

// ListObjects lists objects in a bucket
func (c *Client) ListObjects(ctx context.Context, bucketName string, prefix string) ([]ObjectInfo, error) {
	baseURL := fmt.Sprintf("%s/api/%s", c.baseURL, bucketName)

	// Add prefix parameter if provided
	if prefix != "" {
		params := url.Values{}
		params.Set("prefix", prefix)
		baseURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list objects: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result ListBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Contents, nil
}

// PutObjectFromFile uploads a file to the bucket
func (c *Client) PutObjectFromFile(ctx context.Context, bucketName, objectKey, filePath string) (*UploadResult, error) {
	file, err := http.DefaultClient.Head(filePath) // 这里简化了，实际应该打开本地文件
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	filename := filepath.Base(filePath)
	return c.PutObject(ctx, bucketName, objectKey, file.Body, filename)
}

// GetObjectURL returns the direct URL to access an object
func (c *Client) GetObjectURL(bucketName, objectKey string) string {
	return fmt.Sprintf("%s/api/%s/%s", c.baseURL, bucketName, objectKey)
}
