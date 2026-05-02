package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SwaggerHandler struct {
	swaggerYAML []byte
}

func NewSwaggerHandler(yamlContent []byte) *SwaggerHandler {
	return &SwaggerHandler{swaggerYAML: yamlContent}
}

func (h *SwaggerHandler) ServeYAML(c *gin.Context) {
	c.Header("Content-Type", "application/x-yaml")
	c.Data(http.StatusOK, "application/x-yaml", h.swaggerYAML)
}

func (h *SwaggerHandler) ServeJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	// Redirect to YAML since we're using YAML format
	c.Redirect(http.StatusTemporaryRedirect, "/api/docs/swagger.yaml")
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TaskFlow API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
        .swagger-ui .topbar { display: none; }
        .swagger-ui .info .title { font-size: 2rem; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" crossorigin></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/api/docs/swagger.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                persistAuthorization: true,
                displayRequestDuration: true,
                filter: true,
                showExtensions: true,
                showCommonExtensions: true
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`

func (h *SwaggerHandler) ServeUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerUIHTML)
}
