package tfiber

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

func RegisterOrderApiFiberServer(s *Server, srv any) {
	r := s.Group("/")
	r.Post("/index", _OrderApi_QueryItemOrderInfo0_HTTP_Handler(s, srv))
}

func _OrderApi_QueryItemOrderInfo0_HTTP_Handler(s *Server, srv any) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var in map[string]any
		if err := ctx.QueryParser(in); err != nil {
			return err
		}
		if err := ctx.ParamsParser(in); err != nil {
			return err
		}
		if err := ctx.BodyParser(in); err != nil {
			return err
		}
		SetOperation(ctx.UserContext(), "")
		h := s.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}, ctx.UserContext(), ctx.Path())
		reply, err := h(ctx.UserContext(), &in)
		if err != nil {
			return err
		}
		return ctx.JSON(reply)
	}
}
