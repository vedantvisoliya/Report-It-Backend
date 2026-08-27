package storage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reportit-api/internal/config"
	"strings"
)

func UploadImageToSupabase(filebytes []byte, contentType, filename string, config config.Config) (string, error) {
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", config.SupabaseURL, config.SupabaseBucket, filename)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(filebytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("apikey", config.SupabaseServiceKey)
	req.Header.Set("Authorization", "Bearer "+config.SupabaseServiceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", string(body))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", config.SupabaseURL, config.SupabaseBucket, filename)
	return publicURL, nil
}

func DeleteImageFromSupabase(filename string, config config.Config) error {
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", config.SupabaseURL, config.SupabaseBucket, filename)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", config.SupabaseServiceKey)
	req.Header.Set("Authorization", "Bearer "+config.SupabaseServiceKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s", string(body))
	}

	return nil
}

func FilenameFromURL(url string, bucket string) string {
	parts := strings.Split(url, "/"+bucket+"/")
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}
