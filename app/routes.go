package app

import (
	"net/http"
	"time"

	"github.com/aldelucca1/docsend_scraper/store"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// createRouter creates the default application router
func (a *App) registerRoutes(router *gin.Engine) {

	// Register our middleware
	router.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true), gin.Recovery())
	router.StaticFile("", "./public/index.html")
	router.Static("/-", "./public")

	// Create a custom 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	// Setup route group for the API
	api := router.Group("/api")
	api.GET("documents", a.list)
	api.POST("documents", a.generate)
	api.GET("documents/:id", a.get)
	api.GET("documents/:id/download", a.download)
	api.GET("status", a.status)
}

func (a *App) list(c *gin.Context) {

	// Parse the query params
	owner := c.Query("owner")

	// List the documents
	documents, err := a.service.ListDocuments(owner)
	if err != nil {
		a.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, documents)
}

func (a *App) get(c *gin.Context) {

	// Parse the path params
	id := c.Param("id")

	// Get the document
	document, err := a.service.GetDocument(id)
	if err != nil {
		a.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, document)
}

func (a *App) generate(c *gin.Context) {

	// Parse the incomming parameters
	urlStr := c.PostForm("source_url")
	owner := c.PostForm("owner")
	passcode := c.PostForm("passcode")

	// Generate the document
	document, err := a.service.GenerateDocument(urlStr, owner, passcode)
	if err != nil {
		a.handleError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, document)
}

func (a *App) download(c *gin.Context) {

	// Parse the path params
	id := c.Param("id")

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="` + id + `.pdf"`,
	}

	reader, err := a.service.DownloadDocument(id)
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Render(http.StatusOK, Reader{
		Headers:     extraHeaders,
		ContentType: "application/pdf",
		Reader:      reader,
	})
}

func (a *App) status(c *gin.Context) {

	// Parse the incomming params
	owner := c.Query("owner")

	// Get a new websocket Handler
	handler := websocket.Handler(func(conn *websocket.Conn) {
		a.service.AddClientConnection(owner, conn)
	})
	handler.ServeHTTP(c.Writer, c.Request)
}

func (a *App) handleError(c *gin.Context, err error) {
	if err == store.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	} else if err == store.ErrDuplicateKey {
		c.JSON(http.StatusConflict, gin.H{"code": "CONFLICT", "message": "Document already exists"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
	}
}
