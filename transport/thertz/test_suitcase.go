package thertz

import (
	"context"
)

func RegisterHelloWorldHertzServer(s *Server, srv any) {
	router := s.Router()
	router.POST("/", _HelloWorld_Greeting_Hertz_Handler(s, srv))
}

func _HelloWorld_Greeting_Hertz_Handler(s *Server, srv any) Handler {
	return func(c context.Context, ctx *ReqCtx) {
		var in any
		ctx.Bind(&in)
		ctx.BindQuery(&in)
		ctx.BindPath(&in)
		ctx.BindForm(&in)
		SetOperation(c, "")
		h := s.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		}, c, string(ctx.Path()))
		reply, err := h(c, &in)
		if err != nil {
			panic(err)
		}
		s.Write(ctx, reply)
	}
}
