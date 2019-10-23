package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/jdkato/prose.v2"
)

// A Token represents an individual Token of Text such as a word or punctuation symbol.
// IOB format (short for inside, outside, beginning) is a common tagging format
type Token struct {
	Tag   string `json:"tag"`   // The Token's part-of-speech Tag.
	Text  string `json:"text"`  // The Token's actual content.
	Label string `json:"label"` // The Token's IOB Label.
}

// An Entity represents an individual named-entity.
type Entity struct {
	Text  string `json:"text"`  // The entity's actual content.
	Label string `json:"label"` // The entity's label.
}

var (
	portClient = getEnv("PROSE_PORT", "8082")
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().RequestURI, "/health") {
				return true
			}
			return false
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			return key == os.Getenv("AUTH_KEY"), nil
		},
	}))

	// Routes
	e.GET("/health", getHealth)
	e.POST("/tokens", getTokens)
	e.POST("/entities", getEntities)

	// Start server
	e.Logger.Fatal(e.Start(portClient))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getHealth(c echo.Context) error {
	var response interface{}
	err := json.Unmarshal([]byte(`{"status":"UP"}`), &response)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response)
}

func getTokens(c echo.Context) error {
	var tokens []Token
	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, nil)
	} else {
		text := jsonMap["text"]
		doc, err := prose.NewDocument(text.(string))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		for _, docToken := range doc.Tokens() {
			tokens = append(tokens, Token{
				Tag:   docToken.Tag,
				Text:  docToken.Text,
				Label: docToken.Label,
			})
		}
	}
	return c.JSON(http.StatusOK, tokens)
}

func getEntities(c echo.Context) error {
	var entities []Entity
	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, nil)
	} else {
		text := jsonMap["text"]
		doc, err := prose.NewDocument(text.(string))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}
		for _, docEntities := range doc.Entities() {
			entities = append(entities, Entity{
				Text:  docEntities.Text,
				Label: docEntities.Label,
			})
		}
	}

	return c.JSON(http.StatusOK, entities)
}
