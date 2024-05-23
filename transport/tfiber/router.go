package tfiber

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/gofiber/fiber/v2"
)

type HandlerFunc func(ctx fiber.Ctx) error

type Router struct {
	prefix  string
	srv     *Server
	midList middleware.Middleware
}

func (r *Router) Handle(method, relativePath string, h HandlerFunc) *Router {

	return r

}
