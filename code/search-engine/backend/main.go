package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const movieIndex = "movies"

// Movie represents the schema stored in Elasticsearch.
type Movie struct {
	ID          string  `json:"id"`
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Genre       string  `json:"genre"`
	Rating      float64 `json:"rating"`
	ReleaseYear int     `json:"release_year"`
}

// Pagination metadata returned to the UI.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalHits  int `json:"total_hits"`
	TotalPages int `json:"total_pages"`
}

func main() {
	es := mustCreateElasticsearchClient()
	if err := bootstrapElasticsearch(es); err != nil {
		log.Fatalf("failed to bootstrap Elasticsearch: %v", err)
	}

	router := gin.Default()
	router.Use(corsMiddleware())

	api := router.Group("/api")
	{
		api.GET("/movies", handleSearchMovies(es))
		api.GET("/movies/:id", handleGetMovie(es))
		api.POST("/movies", handleCreateMovie(es))
		api.PUT("/movies/:id", handleUpdateMovie(es))
		api.DELETE("/movies/:id", handleDeleteMovie(es))
	}

	// Serve the static frontend from ../frontend by default.
	frontendDir := getenv("FRONTEND_DIR", "../frontend")
	absDir, err := filepath.Abs(frontendDir)
	if err != nil {
		log.Fatalf("unable to resolve frontend directory: %v", err)
	}
	if _, err := os.Stat(absDir); err == nil {
		router.Static("/", absDir)
	} else {
		log.Printf("frontend directory not found at %s, API will still be available", absDir)
	}

	port := getenv("PORT", "8080")
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func mustCreateElasticsearchClient() *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{getenv("ELASTICSEARCH_ADDRESS", "http://localhost:9200")},
		Username:  os.Getenv("ELASTICSEARCH_USERNAME"),
		Password:  os.Getenv("ELASTICSEARCH_PASSWORD"),
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("unable to create elasticsearch client: %v", err)
	}
	return client
}

func bootstrapElasticsearch(es *elasticsearch.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := es.Indices.Exists([]string{movieIndex}, es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("check index exists: %w", err)
	}
	if exists.StatusCode == http.StatusNotFound {
		if err := createMovieIndex(es); err != nil {
			return err
		}
	}

	return seedMovies(es)
}

func createMovieIndex(es *elasticsearch.Client) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title":        map[string]interface{}{"type": "text"},
				"description":  map[string]interface{}{"type": "text"},
				"genre":        map[string]interface{}{"type": "keyword"},
				"rating":       map[string]interface{}{"type": "float"},
				"release_year": map[string]interface{}{"type": "integer"},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("encode mapping: %w", err)
	}

	res, err := es.Indices.Create(movieIndex, es.Indices.Create.WithBody(&buf))
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index response error: %s", res.String())
	}

	return nil
}

func seedMovies(es *elasticsearch.Client) error {
	res, err := es.Count(es.Count.WithIndex(movieIndex))
	if err != nil {
		return fmt.Errorf("count documents: %w", err)
	}
	defer res.Body.Close()

	var countResponse struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&countResponse); err != nil {
		return fmt.Errorf("decode count response: %w", err)
	}
	if countResponse.Count > 0 {
		return nil
	}

	seedData := []Movie{
		{Title: "Inception", Description: "A thief who steals corporate secrets through dream-sharing technology.", Genre: "Sci-Fi", Rating: 8.8, ReleaseYear: 2010},
		{Title: "The Dark Knight", Description: "Batman battles the Joker in Gotham City.", Genre: "Action", Rating: 9.0, ReleaseYear: 2008},
		{Title: "Interstellar", Description: "Explorers travel through a wormhole in space in an attempt to ensure humanity's survival.", Genre: "Sci-Fi", Rating: 8.6, ReleaseYear: 2014},
		{Title: "La La Land", Description: "A jazz pianist falls for an aspiring actress in Los Angeles.", Genre: "Musical", Rating: 8.0, ReleaseYear: 2016},
		{Title: "The Godfather", Description: "The aging patriarch of an organized crime dynasty transfers control to his reluctant son.", Genre: "Crime", Rating: 9.2, ReleaseYear: 1972},
	}

	for _, movie := range seedData {
		movie.ID = uuid.NewString()
		if err := indexMovie(es, movie.ID, movie); err != nil {
			return fmt.Errorf("seed movie %s: %w", movie.Title, err)
		}
	}

	return nil
}

func handleSearchMovies(es *elasticsearch.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		page := parseIntWithDefault(c.Query("page"), 1)
		pageSize := parseIntWithDefault(c.Query("pageSize"), 5)
		if page < 1 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 50 {
			pageSize = 5
		}

		from := (page - 1) * pageSize

		body := map[string]interface{}{
			"from": from,
			"size": pageSize,
			"sort": []map[string]interface{}{
				{"rating": map[string]interface{}{"order": "desc"}},
			},
		}

		if query == "" {
			body["query"] = map[string]interface{}{"match_all": map[string]interface{}{}}
		} else {
			body["query"] = map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"title^2", "description", "genre"},
				},
			}
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode search query"})
			return
		}

		res, err := es.Search(
			es.Search.WithContext(c.Request.Context()),
			es.Search.WithIndex(movieIndex),
			es.Search.WithBody(&buf),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search request failed"})
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search returned an error"})
			return
		}

		var searchResult struct {
			Hits struct {
				Total struct {
					Value int `json:"value"`
				} `json:"total"`
				Hits []struct {
					ID     string                 `json:"_id"`
					Source map[string]interface{} `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}

		if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode search results"})
			return
		}

		movies := make([]Movie, 0, len(searchResult.Hits.Hits))
		for _, hit := range searchResult.Hits.Hits {
			movie := mapToMovie(hit.Source)
			movie.ID = hit.ID
			movies = append(movies, movie)
		}

		totalHits := searchResult.Hits.Total.Value
		totalPages := (totalHits + pageSize - 1) / pageSize

		c.JSON(http.StatusOK, gin.H{
			"movies": movies,
			"pagination": Pagination{
				Page:       page,
				PageSize:   pageSize,
				TotalHits:  totalHits,
				TotalPages: totalPages,
			},
		})
	}
}

func handleGetMovie(es *elasticsearch.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		res, err := es.Get(movieIndex, id, es.Get.WithContext(c.Request.Context()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movie"})
			return
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		if res.IsError() {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movie"})
			return
		}

		var getResponse struct {
			Source map[string]interface{} `json:"_source"`
		}
		if err := json.NewDecoder(res.Body).Decode(&getResponse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode response"})
			return
		}

		movie := mapToMovie(getResponse.Source)
		movie.ID = id
		c.JSON(http.StatusOK, movie)
	}
}

func handleCreateMovie(es *elasticsearch.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input Movie
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input.ID = uuid.NewString()
		if err := indexMovie(es, input.ID, input); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create movie"})
			return
		}

		c.JSON(http.StatusCreated, input)
	}
}

func handleUpdateMovie(es *elasticsearch.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input Movie
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input.ID = id
		if err := indexMovie(es, id, input); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update movie"})
			return
		}

		c.JSON(http.StatusOK, input)
	}
}

func handleDeleteMovie(es *elasticsearch.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		res, err := es.Delete(movieIndex, id, es.Delete.WithContext(c.Request.Context()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete movie"})
			return
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		if res.IsError() {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete movie"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func indexMovie(es *elasticsearch.Client, id string, movie Movie) error {
	movieJSON := map[string]interface{}{
		"title":        movie.Title,
		"description":  movie.Description,
		"genre":        movie.Genre,
		"rating":       movie.Rating,
		"release_year": movie.ReleaseYear,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(movieJSON); err != nil {
		return fmt.Errorf("encode movie: %w", err)
	}

	res, err := es.Index(
		movieIndex,
		&buf,
		es.Index.WithDocumentID(id),
		es.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("index movie: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index response error: %s", res.String())
	}

	return nil
}

func mapToMovie(source map[string]interface{}) Movie {
	movie := Movie{}
	if title, ok := source["title"].(string); ok {
		movie.Title = title
	}
	if description, ok := source["description"].(string); ok {
		movie.Description = description
	}
	if genre, ok := source["genre"].(string); ok {
		movie.Genre = genre
	}
	if rating, ok := source["rating"].(float64); ok {
		movie.Rating = rating
	} else if ratingNum, ok := source["rating"].(json.Number); ok {
		if value, err := ratingNum.Float64(); err == nil {
			movie.Rating = value
		}
	}
	switch v := source["release_year"].(type) {
	case float64:
		movie.ReleaseYear = int(v)
	case json.Number:
		if value, err := v.Int64(); err == nil {
			movie.ReleaseYear = int(value)
		}
	}
	return movie
}

func parseIntWithDefault(value string, def int) int {
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
