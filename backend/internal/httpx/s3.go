package httpx

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *Server) s3PublicURL(settings map[string]string, key string) string {
	if base := strings.TrimRight(settings["s3PublicBase"], "/"); base != "" {
		return base + "/" + key
	}
	endpoint := strings.TrimRight(settings["s3Endpoint"], "/")
	bucket := settings["s3Bucket"]
	if settings["s3PathStyle"] == "false" {
		if u, err := url.Parse(endpoint); err == nil {
			u.Host = bucket + "." + u.Host
			u.Path = "/" + key
			return u.String()
		}
	}
	return endpoint + "/" + bucket + "/" + key
}

func (s *Server) testS3(w http.ResponseWriter, r *http.Request) {
	if _, err := s.currentUser(r); err != nil {
		writeError(w, 401, "未登录")
		return
	}
	settings, err := s.store.ListSettings()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	key := strings.Trim(settings["s3Prefix"], "/")
	if key != "" {
		key += "/"
	}
	key += "test/connection-check.txt"
	payload := []byte("biu-panel s3 connectivity check\n")
	if err := s.s3PutObject(settings, key, "text/plain; charset=utf-8", payload); err != nil {
		writeError(w, 502, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"key": key, "url": s.s3PublicURL(settings, key), "size": len(payload)})
}

func (s *Server) s3PutObject(settings map[string]string, key, contentType string, payload []byte) error {
	endpoint := strings.TrimRight(settings["s3Endpoint"], "/")
	bucket := settings["s3Bucket"]
	accessKey := settings["s3AccessKey"]
	secretKey := settings["s3SecretKey"]
	region := settings["s3Region"]
	if region == "" {
		region = "auto"
	}
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return errors.New("S3 配置不完整")
	}
	base, err := url.Parse(endpoint)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return errors.New("S3 Endpoint 格式错误")
	}
	pathStyle := settings["s3PathStyle"] != "false"
	objectPath := "/" + bucket + "/" + key
	if !pathStyle {
		base.Host = bucket + "." + base.Host
		objectPath = "/" + key
	}
	base.Path = objectPath
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	payloadHash := sha256Hex(payload)
	headers := map[string]string{
		"host":                 base.Host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
		"content-type":         contentType,
	}
	canonicalHeaders := "content-type:" + headers["content-type"] + "\n" + "host:" + headers["host"] + "\n" + "x-amz-content-sha256:" + payloadHash + "\n" + "x-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := strings.Join([]string{"PUT", uriEncodePath(objectPath), "", canonicalHeaders, signedHeaders, payloadHash}, "\n")
	credentialScope := dateStamp + "/" + region + "/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + credentialScope + "\n" + sha256Hex([]byte(canonicalRequest))
	signature := hex.EncodeToString(hmacSHA256(signingKey(secretKey, dateStamp, region), []byte(stringToSign)))
	authorization := "AWS4-HMAC-SHA256 Credential=" + accessKey + "/" + credentialScope + ", SignedHeaders=" + signedHeaders + ", Signature=" + signature
	req, err := http.NewRequest(http.MethodPut, base.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	req.Header.Set("Authorization", authorization)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("S3 上传失败：%s %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func signingKey(secret, date, region string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func uriEncodePath(v string) string {
	parts := strings.Split(v, "/")
	for i, part := range parts {
		parts[i] = strings.ReplaceAll(url.QueryEscape(part), "+", "%20")
	}
	return strings.Join(parts, "/")
}
