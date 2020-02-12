// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: RESTful Go implementation of gopkg.in/jdkato/prose.v2 package
//          for text processing, including tokenization, part-of-speech tagging, and named-entity extraction
//          by https://github.com/jdkato/prose/tree/v2

package main

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"gopkg.in/jdkato/prose.v2"
	"net/http"
	"os"
	"strings"
	"time"
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
	serverPort = ":" + getEnv("PROSE_PORT", "8080")
	apiKey     = getEnv("API_KEY", "")
	log        = logrus.New()

	// Echo instance
	e = echo.New()
)

func init() {
	log.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "header:X-API-Key",
		Skipper: func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().RequestURI, "/health") {
				return true
			}
			return false
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			log.Debugf("API_KEY: %v", apiKey)
			return key == apiKey, nil
		},
	}))

	// Routes
	e.GET("/health", getHealth)
	e.POST("/tokens", getTokens)
	e.POST("/entities", getEntities)

	// Start server
	e.Logger.Fatal(e.Start(serverPort))
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
		log.Errorf("json.Unmarshal Error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, response)
}

func getTokens(c echo.Context) error {
	var tokens []Token
	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		log.Errorf("json.NewDecoder Error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	} else {
		text := jsonMap["text"]
		doc, err := prose.NewDocument(text.(string))
		if err != nil {
			log.Errorf("prose.NewDocument Error: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err)
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
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	} else {
		text := jsonMap["text"]
		doc, err := prose.NewDocument(text.(string))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
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
