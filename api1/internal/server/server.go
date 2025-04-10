package server

import (
	"github.com/gin-gonic/gin"
	"github.com/nikita89756/testEffectiveMobile/internal/handlers"
	middleware "github.com/nikita89756/testEffectiveMobile/internal/middlware"
)

type Server struct {
	Host    string
	Port    string
	Handler *handlers.Handler
}

func New(Host string, Port string, handlers *handlers.Handler) *Server {
	return &Server{
		Host:    Host,
		Port:    Port,
		Handler: handlers,
	}
	
}

func (s *Server) CreateRoute() *gin.Engine{
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(gin.ErrorLogger())
	router.Use(middleware.CORSMiddleware())
	api := router.Group("/api")
	{
		api.GET("/persons", s.Handler.GetPersons)
		api.GET("/persons/:id", s.Handler.FindPersonByID)
		api.POST("/persons", s.Handler.CreatePerson)
		api.DELETE("/persons/:id", s.Handler.DeletePersonByID)
		api.PUT("/persons/:id", s.Handler.UpdatePersonByID)
	}

	return router
	
}