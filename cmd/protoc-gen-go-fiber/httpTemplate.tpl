{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{- range .MethodSets}}
const FiberOperation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

type {{.ServiceType}}FiberServer interface {
{{- range .MethodSets}}
	{{- if ne .Comment ""}}
	{{.Comment}}
	{{- end}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}FiberServer(s *tfiber.Server, srv {{.ServiceType}}FiberServer) {
	r := s.Group("/")
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_Fiber_Handler(s, srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_Fiber_Handler(s *tfiber.Server, srv {{$svrType}}FiberServer) tfiber.Handler {
	return func(ctx *tfiber.Ctx) error {
		var in {{.Request}}
		{{- if .HasBody}}
		if err := ctx.BodyParser(&in{{.Body}}); err != nil {
			return err
		}
		{{- end}}
		if err := ctx.QueryParser(&in); err != nil {
			return err
		}
		{{- if .HasVars}}
		if err := ctx.ParamsParser(&in); err != nil {
			return err
		}
		{{- end}}
		tfiber.SetOperation(ctx.UserContext(),FiberOperation{{$svrType}}{{.OriginalName}})
		h := s.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		}, ctx.UserContext(), ctx.Path())
		out, err := h(ctx.UserContext(), &in)
		if err != nil {
			return err
		}
		reply := out.(*{{.Reply}})
		return s.Write(ctx, reply)
	}
}
{{end}}
