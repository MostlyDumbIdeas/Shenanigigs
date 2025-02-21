package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"shenanigigs/common/telemetry"
	"shenanigigs/ingestion/internal/config"
	"shenanigigs/ingestion/internal/errors"
	"shenanigigs/ingestion/internal/models"

	"go.uber.org/zap"
)

var tracer = telemetry.GetTracer("shenanigigs/ingestion/api")

type JobSourceClient interface {
	GetItem(ctx context.Context, id int) (*models.SourcePost, error)
	GetTopStories(ctx context.Context) ([]int, error)
	SearchHiringThreads(ctx context.Context) ([]int, error)
}

type jobSourceClient struct {
	client *http.Client
	logger *zap.Logger
	config *config.Config
}

func (c *jobSourceClient) SearchHiringThreads(ctx context.Context) ([]int, error) {
	ctx, span := tracer.Start(ctx, "SearchHiringThreads")
	defer span.End()

	url := fmt.Sprintf("%s/search?tags=story,author_whoishiring&query=Ask+HN:+Who+is+hiring?&numericFilters=created_at_i>%d",
		c.config.HNSearchAPIBaseURL,
		time.Now().AddDate(0, -6, 0).Unix())
	c.logger.Debug("searching for hiring threads", zap.String("url", url))
	span.SetAttributes(telemetry.String("http.url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, errors.Internal("creating request", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		c.logger.Error("failed to execute request", zap.Error(err))
		return nil, errors.Internal("executing request", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.logger.Warn("failed to close response body", zap.Error(cerr))
		}
	}()

	span.SetAttributes(
		telemetry.Int("http.status_code", resp.StatusCode),
		telemetry.String("http.method", http.MethodGet),
	)

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Int("status_code", resp.StatusCode))
		return nil, errors.Internal(fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var searchResult struct {
		Hits []struct {
			ObjectID string `json:"objectID"`
			Title    string `json:"title"`
			Author   string `json:"author"`
		} `json:"hits"`
		NbHits int `json:"nbHits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		c.logger.Error("failed to decode response", zap.Error(err))
		return nil, errors.Internal("decoding response", err)
	}

	c.logger.Info("search response stats",
		zap.Int("total_hits", searchResult.NbHits))

	ids := make([]int, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		id, err := strconv.Atoi(hit.ObjectID)
		if err != nil {
			c.logger.Warn("invalid story ID",
				zap.String("id", hit.ObjectID),
				zap.String("title", hit.Title),
				zap.String("author", hit.Author))
			continue
		}
		ids = append(ids, id)
	}

	c.logger.Debug("successfully fetched hiring threads",
		zap.Int("count", len(ids)))
	return ids, nil
}

func NewJobSourceClient(logger *zap.Logger, config *config.Config) JobSourceClient {
	return &jobSourceClient{
		client: &http.Client{
			Timeout: config.HNAPITimeout,
		},
		logger: logger,
		config: config,
	}
}

func (c *jobSourceClient) GetItem(ctx context.Context, id int) (*models.SourcePost, error) {
	url := fmt.Sprintf("%s/item/%d.json", c.config.HNAPIBaseURL, id)
	c.logger.Debug("fetching item", zap.Int("id", id), zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Internal("creating request", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("failed to execute request", zap.Int("id", id), zap.Error(err))
		return nil, errors.Internal("executing request", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.logger.Warn("failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		c.logger.Warn("item not found", zap.Int("id", id))
		return nil, errors.NotFound("item not found", nil)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code",
			zap.Int("id", id),
			zap.Int("status_code", resp.StatusCode))
		return nil, errors.Internal(fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var post models.SourcePost
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		c.logger.Error("failed to decode response", zap.Int("id", id), zap.Error(err))
		return nil, errors.Internal("decoding response", err)
	}

	c.logger.Debug("successfully fetched item",
		zap.Int("id", id),
		zap.String("title", post.Title))

	return &post, nil
}

func (c *jobSourceClient) GetTopStories(ctx context.Context) ([]int, error) {
	url := fmt.Sprintf("%s/topstories.json", c.config.HNAPIBaseURL)
	c.logger.Debug("fetching top stories", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Internal("creating request", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("failed to execute request", zap.Error(err))
		return nil, errors.Internal("executing request", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.logger.Warn("failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Int("status_code", resp.StatusCode))
		return nil, errors.Internal(fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		c.logger.Error("failed to decode response", zap.Error(err))
		return nil, errors.Internal("decoding response", err)
	}

	c.logger.Debug("successfully fetched top stories", zap.Int("count", len(ids)))
	return ids, nil
}
