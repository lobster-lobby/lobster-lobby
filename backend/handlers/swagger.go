package handlers

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

//go:embed docs/openapi.yaml
var openAPISpec []byte

// SwaggerDocs serves the Swagger UI and OpenAPI spec
func SwaggerDocs() gin.HandlerFunc {
	// Create sub-filesystem for swagger-ui
	subFS, err := fs.Sub(swaggerUI, "swagger-ui")
	if err != nil {
		log.Fatalf("Failed to create swagger-ui sub-filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	return func(c *gin.Context) {
		path := c.Param("filepath")

		// Serve embedded OpenAPI spec
		if path == "/openapi.yaml" || path == "openapi.yaml" {
			c.Data(http.StatusOK, "application/yaml", openAPISpec)
			return
		}

		// Remove leading slash for file serving
		path = strings.TrimPrefix(path, "/")

		// Default to index.html
		if path == "" || path == "/" {
			path = "index.html"
		}

		// Serve from embedded filesystem using a cloned request to avoid mutating the original
		req := c.Request.Clone(c.Request.Context())
		req.URL.Path = "/" + path
		fileServer.ServeHTTP(c.Writer, req)
	}
}
