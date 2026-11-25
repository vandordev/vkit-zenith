// Package storage provides private S3-compatible object storage with presigned access.
package storage

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Settings struct {
	Endpoint, Region, Bucket, AccessKeyID, SecretAccessKey string
	PresignTTL                                             time.Duration
}

type Object struct {
	Size        int64
	ContentType string
}

type Client struct {
	bucket  string
	ttl     time.Duration
	s3      *s3.Client
	presign *s3.PresignClient
}

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9/_.-]{1,512}$`)

func New(s Settings) (*Client, error) {
	u, err := url.Parse(s.Endpoint)
	hostname := ""
	if err == nil {
		hostname = u.Hostname()
	}
	if err != nil || hostname == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(s.Endpoint, "#") || (u.Scheme != "https" && !(u.Scheme == "http" && loopbackHost(hostname))) {
		return nil, errors.New("invalid storage endpoint")
	}
	if s.Region == "" || s.Bucket == "" || s.AccessKeyID == "" || s.SecretAccessKey == "" {
		return nil, errors.New("invalid storage settings")
	}
	if s.PresignTTL < time.Minute || s.PresignTTL > time.Hour {
		return nil, errors.New("invalid storage presign ttl")
	}
	client := s3.New(s3.Options{
		Region:       s.Region,
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(s.AccessKeyID, s.SecretAccessKey, "")),
		BaseEndpoint: aws.String(strings.TrimSuffix(s.Endpoint, "/")),
		UsePathStyle: true,
	})
	return &Client{bucket: s.Bucket, ttl: s.PresignTTL, s3: client, presign: s3.NewPresignClient(client)}, nil
}

func loopbackHost(host string) bool {
	return strings.EqualFold(host, "localhost") || (net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback())
}

func (c *Client) PresignPut(ctx context.Context, key, contentType string) (string, http.Header, error) {
	if !validKey(key) {
		return "", nil, errors.New("invalid storage key")
	}
	if !allowedForKey(key, contentType) {
		return "", nil, errors.New("unsupported storage content type")
	}
	r, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key), ContentType: aws.String(contentType)}, s3.WithPresignExpires(c.ttl))
	if err != nil {
		return "", nil, errors.New("could not presign storage upload")
	}
	h := make(http.Header)
	h.Set("Content-Type", contentType)
	return r.URL, h, nil
}

func (c *Client) PresignGet(ctx context.Context, key string) (string, error) {
	if !validKey(key) {
		return "", errors.New("invalid storage key")
	}
	r, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key)}, s3.WithPresignExpires(c.ttl))
	if err != nil {
		return "", errors.New("could not presign storage download")
	}
	return r.URL, nil
}

func (c *Client) Head(ctx context.Context, key string) (Object, error) {
	if !validKey(key) {
		return Object{}, errors.New("invalid storage key")
	}
	r, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key)})
	if err != nil {
		return Object{}, errors.New("could not inspect stored object")
	}
	var result Object
	if r.ContentLength != nil {
		result.Size = *r.ContentLength
	}
	if r.ContentType != nil {
		result.ContentType = *r.ContentType
	}
	return result, nil
}

func LogoKey(organizationID uuid.UUID, contentType string) (string, error) {
	ext, ok := extension(contentType)
	if !ok {
		return "", fmt.Errorf("unsupported logo content type")
	}
	return fmt.Sprintf("orgs/%s/logo/%s.%s", organizationID, uuid.New(), ext), nil
}

func validKey(key string) bool { return keyPattern.MatchString(key) }
func allowedForKey(key, ct string) bool {
	_, ok := extension(ct)
	return ok && strings.Contains(key, "/logo/")
}
func extension(ct string) (string, bool) {
	switch ct {
	case "image/png":
		return "png", true
	case "image/jpeg":
		return "jpg", true
	case "image/webp":
		return "webp", true
	default:
		return "", false
	}
}
