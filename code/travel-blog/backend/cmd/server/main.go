package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Country struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Places      []Place   `json:"places"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Place struct {
	ID          int64      `json:"id"`
	CountryID   int64      `json:"country_id"`
	Name        string     `json:"name"`
	Category    string     `json:"category"`
	City        string     `json:"city"`
	Description string     `json:"description"`
	VisitedAt   *time.Time `json:"visited_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type App struct {
	db *sql.DB
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	app := &App{db: db}
	if err := app.ensureSchema(); err != nil {
		log.Fatalf("failed to ensure schema: %v", err)
	}

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.GET("/countries", app.listCountries)
		api.POST("/countries", app.createCountry)
		api.GET("/countries/:id", app.getCountry)
		api.PUT("/countries/:id", app.updateCountry)
		api.DELETE("/countries/:id", app.deleteCountry)

		api.POST("/countries/:id/places", app.createPlace)
		api.PUT("/places/:id", app.updatePlace)
		api.DELETE("/places/:id", app.deletePlace)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func (a *App) ensureSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS countries (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT NOT NULL DEFAULT '',
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );`,
		`CREATE TABLE IF NOT EXISTS places (
            id SERIAL PRIMARY KEY,
            country_id INTEGER NOT NULL REFERENCES countries(id) ON DELETE CASCADE,
            name TEXT NOT NULL,
            category TEXT NOT NULL,
            city TEXT NOT NULL DEFAULT '',
            description TEXT NOT NULL DEFAULT '',
            visited_at DATE,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );`,
		`CREATE OR REPLACE FUNCTION set_updated_at()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.updated_at = NOW();
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;`,
		`CREATE OR REPLACE TRIGGER countries_updated_at
        BEFORE UPDATE ON countries
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,
		`CREATE OR REPLACE TRIGGER places_updated_at
        BEFORE UPDATE ON places
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,
	}

	for _, q := range queries {
		if _, err := a.db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) listCountries(c *gin.Context) {
	countries, err := a.fetchCountries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, countries)
}

func (a *App) fetchCountries() ([]Country, error) {
	rows, err := a.db.Query(`SELECT id, name, description, created_at, updated_at FROM countries ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.ID, &country.Name, &country.Description, &country.CreatedAt, &country.UpdatedAt); err != nil {
			return nil, err
		}
		places, err := a.fetchPlaces(country.ID)
		if err != nil {
			return nil, err
		}
		country.Places = places
		countries = append(countries, country)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return countries, nil
}

func (a *App) fetchCountry(id int64) (*Country, error) {
	var country Country
	err := a.db.QueryRow(`SELECT id, name, description, created_at, updated_at FROM countries WHERE id=$1`, id).
		Scan(&country.ID, &country.Name, &country.Description, &country.CreatedAt, &country.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	places, err := a.fetchPlaces(id)
	if err != nil {
		return nil, err
	}
	country.Places = places
	return &country, nil
}

func (a *App) fetchPlaces(countryID int64) ([]Place, error) {
	rows, err := a.db.Query(`SELECT id, country_id, name, category, city, description, visited_at, created_at, updated_at FROM places WHERE country_id=$1 ORDER BY visited_at DESC NULLS LAST, name`, countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []Place
	for rows.Next() {
		var place Place
		if err := rows.Scan(&place.ID, &place.CountryID, &place.Name, &place.Category, &place.City, &place.Description, &place.VisitedAt, &place.CreatedAt, &place.UpdatedAt); err != nil {
			return nil, err
		}
		places = append(places, place)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return places, nil
}

func (a *App) createCountry(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
		return
	}

	description := strings.TrimSpace(input.Description)

	var id int64
	err := a.db.QueryRow(`INSERT INTO countries(name, description) VALUES($1, $2) RETURNING id`, name, description).
		Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	country, err := a.fetchCountry(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, country)
}

func (a *App) getCountry(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	country, err := a.fetchCountry(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if country == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "country not found"})
		return
	}

	c.JSON(http.StatusOK, country)
}

func (a *App) updateCountry(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var name interface{}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed != "" {
			name = trimmed
		} else {
			name = ""
		}
	}

	var description interface{}
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}

	res, err := a.db.Exec(`UPDATE countries SET name = COALESCE($1, name), description = COALESCE($2, description) WHERE id=$3`, name, description, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "country not found"})
		return
	}

	country, err := a.fetchCountry(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, country)
}

func (a *App) deleteCountry(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := a.db.Exec(`DELETE FROM countries WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "country not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *App) createPlace(c *gin.Context) {
	countryID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		Name        string  `json:"name" binding:"required"`
		Category    string  `json:"category" binding:"required"`
		City        string  `json:"city"`
		Description string  `json:"description"`
		VisitedAt   *string `json:"visited_at"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(input.Name)
	category := strings.TrimSpace(input.Category)
	city := strings.TrimSpace(input.City)
	description := strings.TrimSpace(input.Description)

	if name == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and category are required"})
		return
	}

	var visitedAt *time.Time
	if input.VisitedAt != nil && *input.VisitedAt != "" {
		t, err := time.Parse("2006-01-02", *input.VisitedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visited_at format, expected YYYY-MM-DD"})
			return
		}
		visitedAt = &t
	}

	var id int64
	err = a.db.QueryRow(`INSERT INTO places(country_id, name, category, city, description, visited_at) VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		countryID, name, category, city, description, visitedAt).
		Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	country, err := a.fetchCountry(countryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, country)
}

func (a *App) updatePlace(c *gin.Context) {
	placeID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		Name        *string `json:"name"`
		Category    *string `json:"category"`
		City        *string `json:"city"`
		Description *string `json:"description"`
		VisitedAt   *string `json:"visited_at"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setVisited := false
	var visitedAt interface{}
	if input.VisitedAt != nil {
		setVisited = true
		if *input.VisitedAt == "" {
			visitedAt = nil
		} else {
			t, err := time.Parse("2006-01-02", *input.VisitedAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visited_at format, expected YYYY-MM-DD"})
				return
			}
			visitedAt = t
		}
	}

	var name interface{}
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}
	var category interface{}
	if input.Category != nil {
		category = strings.TrimSpace(*input.Category)
	}
	var city interface{}
	if input.City != nil {
		city = strings.TrimSpace(*input.City)
	}
	var description interface{}
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}

	res, err := a.db.Exec(`UPDATE places SET
        name = COALESCE($1, name),
        category = COALESCE($2, category),
        city = COALESCE($3, city),
        description = COALESCE($4, description),
        visited_at = CASE WHEN $5 THEN $6 ELSE visited_at END
    WHERE id=$7`, name, category, city, description, setVisited, visitedAt, placeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "place not found"})
		return
	}

	var countryID int64
	err = a.db.QueryRow(`SELECT country_id FROM places WHERE id=$1`, placeID).Scan(&countryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	country, err := a.fetchCountry(countryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, country)
}

func (a *App) deletePlace(c *gin.Context) {
	placeID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var countryID int64
	if err := a.db.QueryRow(`SELECT country_id FROM places WHERE id=$1`, placeID).Scan(&countryID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "place not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := a.db.Exec(`DELETE FROM places WHERE id=$1`, placeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "place not found"})
		return
	}

	country, err := a.fetchCountry(countryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, country)
}

func parseIDParam(c *gin.Context, name string) (int64, error) {
	idStr := c.Param(name)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
