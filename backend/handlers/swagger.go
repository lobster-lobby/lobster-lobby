package handlers

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

// SwaggerDocs serves the Swagger UI and OpenAPI spec
func SwaggerDocs(specPath string) gin.HandlerFunc {
	// Create sub-filesystem for swagger-ui
	subFS, _ := fs.Sub(swaggerUI, "swagger-ui")
	fileServer := http.FileServer(http.FS(subFS))

	return func(c *gin.Context) {
		path := c.Param("filepath")

		// Serve OpenAPI spec
		if path == "/openapi.yaml" || path == "openapi.yaml" {
			c.File(specPath)
			return
		}

		// Remove leading slash for file serving
		path = strings.TrimPrefix(path, "/")

		// Default to index.html
		if path == "" || path == "/" {
			path = "index.html"
		}

		// Serve from embedded filesystem
		c.Request.URL.Path = "/" + path
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
