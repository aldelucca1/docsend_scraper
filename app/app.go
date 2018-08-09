package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aldelucca1/docsend_scraper/service"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
)

// App is our main application
type App struct {
	router  *gin.Engine
	server  *http.Server
	service *service.Service
}

// NewApp creates a new App instance
func NewApp() *App {
	app := new(App)
	app.service = service.NewService()
	app.router = gin.New()
	app.registerRoutes(app.router)
	return app
}

// Run starts our app listening for requests
func (a *App) Run() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := a.Start(); err != nil {
			if err != http.ErrServerClosed {
				logger.Error(err)
				stop <- os.Kill
			}
		}
	}()

	<-stop

	a.Shutdown()
}

// Start - Starts the Server and listens for requests
func (a *App) Start() error {
	if err := a.service.Start(); err != nil {
		return err
	}
	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: a.router,
	}
	logger.Infof("Listening for HTTP on '%s'", a.server.Addr)
	return a.server.ListenAndServe()
}

// Shutdown - Gracefully shuts down the Server instance
func (a *App) Shutdown() {
	logger.Info("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.server.Shutdown(ctx)
	a.service.Stop()
}
