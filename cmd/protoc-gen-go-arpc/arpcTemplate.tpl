{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{- range .MethodSets}}
const ArpcOperation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

type {{.ServiceType}}ArpcServer interface {
{{- range .MethodSets}}
	{{- if ne .Comment ""}}
	{{.Comment}}
	{{- end}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}ArpcServer(s *tarpc.Server, srv {{.ServiceType}}ArpcServer) {
	{{- range .Methods}}
	s.Handle(ArpcOperation{{$svrType}}{{.OriginalName}}, _{{$svrType}}_{{.Name}}{{.Num}}_Arpc_Handler(s, srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_Arpc_Handler(s *tarpc.Server, srv {{$svrType}}ArpcServer) tarpc.HandlerFunc {
	return func(c *tarpc.Ctx) {
	    var err error
	    ctx, bytes, err := s.DecodeRequest(c)
	    if err != nil {
	        panic(err)
	    }
	    // timeout
	    ctx, cancel := s.Timeout(ctx)
	    defer cancel()
		var in {{.Request}}
		err = s.DecodeData(bytes, &in)
	    if err != nil {
	        panic(err)
	    }
		tarpc.SetOperation(ctx, ArpcOperation{{$svrType}}{{.OriginalName}})
		h := s.Middleware(ctx, func(_ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(_ctx, req.(*{{.Request}}))
		})
		out, err := h(ctx, &in)
		if err != nil {
			panic(err)
		}
		reply := out.(*{{.Reply}})
		resp := s.EncodeResponse(ctx, reply, err)
		s.Write(c, resp)
	}
}
{{end}}

type {{.ServiceType}}ArpcClient interface {
{{- range .MethodSets}}
	{{- if ne .Comment ""}}
	{{.Comment}}
	{{- end}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

type _{{.ServiceType}}ArpcClientImpl struct {
    cc *tarpc.Client
}

func New{{.ServiceType}}ArpcClient(cc *tarpc.Client) {{.ServiceType}}ArpcClient {
    return &_{{.ServiceType}}ArpcClientImpl{cc}
}

{{range .Methods}}
func (c *_{{$svrType}}ArpcClientImpl) {{.Name}}(ctx context.Context, req *{{.Request}}) (*{{.Reply}}, error) {
    out := new({{.Reply}})
    err := c.cc.Call(ctx, ArpcOperation{{$svrType}}{{.OriginalName}}, req, out)
    if err != nil {
        return nil, err
    }
    return out, nil
}
{{end}}