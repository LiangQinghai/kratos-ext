{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{- range .MethodSets}}
const HertzOperation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

type {{.ServiceType}}HertzServer interface {
{{- range .MethodSets}}
	{{- if ne .Comment ""}}
	{{.Comment}}
	{{- end}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}HertzServer(s *thertz.Server, srv {{.ServiceType}}HertzServer) {
	r := s.Router()
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_Hertz_Handler(s, srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_Hertz_Handler(s *thertz.Server, srv {{$svrType}}HertzServer) thertz.Handler {
	return func(c context.Context, ctx *ReqCtx) {
		var in {{.Request}}
		{{- if .HasBody}}
		if err := ctx.Bind(&in{{.Body}}); err != nil {
			panic(err)
		}
		{{- end}}
		if err := ctx.BindQuery(&in); err != nil {
			panic(err)
		}
		{{- if .HasVars}}
		if err := ctx.BindPath(&in); err != nil {
			panic(err)
		}
		{{- end}}
		thertz.SetOperation(c,HertzOperation{{$svrType}}{{.OriginalName}})
		h := s.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		}, c, string(ctx.Path()))
		out, err := h(ctx.UserContext(), &in)
		if err != nil {
			panic(err)
		}
		reply := out.(*{{.Reply}})
		s.Write(ctx, reply)
	}
}
{{end}}
