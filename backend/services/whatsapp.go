package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const metaAPIBase = "https://graph.facebook.com/v25.0"

// SendTemplateMessage sends a WhatsApp template message.
// components is a JSON-encoded array of Meta template components.
// Optional tenantPhoneNumberID and tenantToken override platform credentials.
func SendTemplateMessage(phone, templateName, language, componentsJSON string, tenantCreds ...string) (string, error) {
	phoneNumberID := os.Getenv("META_PHONE_NUMBER_ID")
	token := os.Getenv("META_ACCESS_TOKEN")
	if len(tenantCreds) >= 2 && tenantCreds[0] != "" && tenantCreds[1] != "" {
		phoneNumberID = tenantCreds[0]
		token = tenantCreds[1]
	}

	var components []interface{}
	json.Unmarshal([]byte(componentsJSON), &components)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                phone,
		"type":              "template",
		"template": map[string]interface{}{
			"name":       templateName,
			"language":   map[string]string{"code": language},
			"components": components,
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/%s/messages", metaAPIBase, phoneNumberID)
	log.Printf("[meta] POST %s  body=%s", url, string(body))
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		var metaErr struct {
			Error struct {
				Message   string `json:"message"`
				Code      int    `json:"code"`
				Type      string `json:"type"`
				ErrorData struct {
					Details string `json:"details"`
				} `json:"error_data"`
				FbtraceID string `json:"fbtrace_id"`
			} `json:"error"`
		}
		json.Unmarshal(rawBody, &metaErr)
		if metaErr.Error.Code != 0 {
			return "", fmt.Errorf("meta %d (%s): %s — %s [trace:%s]",
				metaErr.Error.Code, metaErr.Error.Type,
				metaErr.Error.Message, metaErr.Error.ErrorData.Details,
				metaErr.Error.FbtraceID)
		}
		return "", fmt.Errorf("meta HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var result map[string]interface{}
	json.Unmarshal(rawBody, &result)
	if messages, ok := result["messages"].([]interface{}); ok && len(messages) > 0 {
		if msg, ok := messages[0].(map[string]interface{}); ok {
			if id, ok := msg["id"].(string); ok {
				return id, nil
			}
		}
	}
	return "", nil
}

// SendImageWithToken pushes a standalone WhatsApp image message via public URL.
func SendImageWithToken(phoneNumberID, token, to, imageURL, caption string) (string, error) {
	if phoneNumberID == "" || token == "" {
		return "", fmt.Errorf("meta credentials not configured")
	}
	image := map[string]interface{}{"link": imageURL}
	if caption != "" {
		image["caption"] = caption
	}
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "image",
		"image":             image,
	}
	raw, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/%s/messages", metaAPIBase, phoneNumberID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("meta image send HTTP %d: %s", resp.StatusCode, string(buf))
	}
	var parsed struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	_ = json.Unmarshal(buf, &parsed)
	if len(parsed.Messages) > 0 {
		return parsed.Messages[0].ID, nil
	}
	return "", nil
}

// SendFreeformTextWithToken sends a plain text message within the 24h service window.
func SendFreeformTextWithToken(phoneNumberID, token, to, body string) (string, error) {
	if phoneNumberID == "" || token == "" {
		return "", fmt.Errorf("meta credentials not configured")
	}
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "text",
		"text":              map[string]interface{}{"body": body, "preview_url": false},
	}
	raw, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/%s/messages", metaAPIBase, phoneNumberID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("meta HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	if messages, ok := result["messages"].([]interface{}); ok && len(messages) > 0 {
		if msg, ok := messages[0].(map[string]interface{}); ok {
			if id, ok := msg["id"].(string); ok {
				return id, nil
			}
		}
	}
	return "", nil
}

// SubmitTemplateToMeta submits a template to Meta for approval.
// Returns the Meta template ID on success.
func SubmitTemplateToMeta(wabaID, name, language, category string, components []map[string]interface{}, tenantToken ...string) (string, error) {
	token := os.Getenv("META_ACCESS_TOKEN")
	if len(tenantToken) > 0 && tenantToken[0] != "" {
		token = tenantToken[0]
	}
	apiURL := fmt.Sprintf("%s/%s/message_templates", metaAPIBase, wabaID)

	payload := map[string]interface{}{
		"name":             name,
		"language":         language,
		"category":         category,
		"parameter_format": "NAMED",
		"components":       components,
	}
	body, _ := json.Marshal(payload)
	log.Printf("[meta] POST %s  body=%s", apiURL, string(body))
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		var metaErr struct {
			Error struct {
				Message   string `json:"message"`
				Code      int    `json:"code"`
				Type      string `json:"type"`
				ErrorData struct {
					Details string `json:"details"`
				} `json:"error_data"`
			} `json:"error"`
		}
		json.Unmarshal(rawBody, &metaErr)
		if metaErr.Error.Code != 0 {
			return "", fmt.Errorf("meta %d (%s): %s — %s",
				metaErr.Error.Code, metaErr.Error.Type,
				metaErr.Error.Message, metaErr.Error.ErrorData.Details)
		}
		return "", fmt.Errorf("meta HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var result map[string]interface{}
	json.Unmarshal(rawBody, &result)
	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	return "", nil
}

// UploadImageToMeta uploads an image to Meta's Resumable Upload API.
// Returns the file handle needed for template header_handle.
func UploadImageToMeta(imageURL string, tenantToken ...string) (string, error) {
	token := os.Getenv("META_ACCESS_TOKEN")
	if len(tenantToken) > 0 && tenantToken[0] != "" {
		token = tenantToken[0]
	}
	appID := os.Getenv("META_APP_ID")
	if appID == "" {
		return "", fmt.Errorf("META_APP_ID not configured")
	}

	imgResp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer imgResp.Body.Close()
	imgData, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}
	contentType := imgResp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	// Step 1: Create upload session
	sessionURL := fmt.Sprintf("%s/%s/uploads", metaAPIBase, appID)
	sessionPayload, _ := json.Marshal(map[string]interface{}{
		"file_length": len(imgData),
		"file_type":   contentType,
	})
	req, err := http.NewRequest("POST", sessionURL, bytes.NewReader(sessionPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload session request failed: %w", err)
	}
	defer resp.Body.Close()
	sessBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload session failed HTTP %d: %s", resp.StatusCode, string(sessBody))
	}

	var sessionResult struct {
		ID string `json:"id"`
	}
	json.Unmarshal(sessBody, &sessionResult)
	if sessionResult.ID == "" {
		return "", fmt.Errorf("no upload session ID in response: %s", string(sessBody))
	}

	// Step 2: Upload binary data
	uploadURL := fmt.Sprintf("%s/%s", metaAPIBase, sessionResult.ID)
	req2, err := http.NewRequest("POST", uploadURL, bytes.NewReader(imgData))
	if err != nil {
		return "", err
	}
	req2.Header.Set("Authorization", "OAuth "+token)
	req2.Header.Set("file_offset", "0")
	req2.Header.Set("Content-Type", contentType)

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return "", fmt.Errorf("upload binary failed: %w", err)
	}
	defer resp2.Body.Close()
	uploadBody, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode >= 300 {
		return "", fmt.Errorf("upload binary failed HTTP %d: %s", resp2.StatusCode, string(uploadBody))
	}

	var uploadResult struct {
		Handle string `json:"h"`
	}
	json.Unmarshal(uploadBody, &uploadResult)
	if uploadResult.Handle == "" {
		return "", fmt.Errorf("no handle in upload response: %s", string(uploadBody))
	}
	log.Printf("[meta] uploaded image, handle=%s", uploadResult.Handle[:20]+"...")
	return uploadResult.Handle, nil
}

// SyncTemplatesFromMeta fetches all templates for a WABA and returns them.
func SyncTemplatesFromMeta(wabaID string) ([]map[string]interface{}, error) {
	token := os.Getenv("META_ACCESS_TOKEN")
	url := fmt.Sprintf("%s/%s/message_templates?fields=id,name,language,status,category,components,rejected_reason&limit=100", metaAPIBase, wabaID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("meta HTTP %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(body, &result)
	return result.Data, nil
}
