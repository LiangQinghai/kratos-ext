{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{- range .MethodSets}}
const Operation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

type {{.ServiceType}}HTTPServer interface {
{{- range .MethodSets}}
	{{- if ne .Comment ""}}
	{{.Comment}}
	{{- end}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}HTTPServer(s *tfiber.Server, srv {{.ServiceType}}HTTPServer) {
	r := s.Group("/")
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(s, srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(s *tfiber.Server, srv {{$svrType}}HTTPServer) tfiber.Handler {
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
		tfiber.SetOperation(ctx.UserContext(),Operation{{$svrType}}{{.OriginalName}})
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		})
		out, err := h(ctx.UserContext(), &in)
		if err != nil {
			return err
		}
		reply := out.(*{{.Reply}})
		return ctx.JSON(reply)
	}
}
{{end}}
