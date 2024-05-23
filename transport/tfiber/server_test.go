package tfiber

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gofiber/fiber/v2"
	"io"
	http2 "net/http"
	"strings"
	"testing"
	"time"
)

type testData struct {
	Path string `json:"path"`
}

type bindData struct {
	Path  string `json:"path"`
	Query string `json:"query"`
	Body  string `json:"body"`
}

func newHandleFuncWrapper() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.JSON(testData{Path: string(ctx.Request().URI().Path())})
	}
}

func newBindHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var in bindData
		err := ctx.ParamsParser(&in)
		if err != nil {
			return err
		}
		err = ctx.QueryParser(&in)
		if err != nil {
			return err
		}
		err = ctx.BodyParser(&in)
		if err != nil {
			return err
		}
		return ctx.JSON(&in)
	}
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	srv := NewServer()
	srv.All("/index", newHandleFuncWrapper())
	srv.All("/index/:id<int>", newHandleFuncWrapper())
	srv.Route("/test/prefix", func(router fiber.Router) {
		router.Use(newHandleFuncWrapper())
	})
	srv.All("/bind/:path", newBindHandler())
	srv.Group("/errors").Get("/cause", func(ctx *fiber.Ctx) error {
		return kratoserrors.BadRequest(
			"xxx",
			"zzz",
		).WithMetadata(map[string]string{"foo": "bar"}).
			WithCause(errors.New("error cause"))
	})

	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}

	go func() {
		if err := srv.Start(ctx); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)
	testBind(t, srv)
	testHeader(t, srv)
	testClient(t, srv)
	testAccept(t, srv)
	time.Sleep(time.Second)
	if srv.Stop(ctx) != nil {
		t.Errorf("expected nil got %v", srv.Stop(ctx))
	}

}

func testAccept(t *testing.T, srv *Server) {
	tests := []struct {
		method      string
		path        string
		contentType string
	}{
		{http2.MethodGet, "/errors/cause", "application/json"},
		{http2.MethodGet, "/errors/cause", "application/proto"},
	}
	e, err := srv.Endpoint()
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	client, err := http.NewClient(context.Background(), http.WithEndpoint(e.Host))
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	for _, test := range tests {
		req, err := http2.NewRequest(test.method, e.String()+test.path, nil)
		if err != nil {
			t.Errorf("expected nil got %v", err)
		}
		req.Header.Set("Content-Type", test.contentType)
		resp, err := client.Do(req)
		if kratoserrors.Code(err) != 500 {
			t.Errorf("expected 500 got %v", err)
		}
		if err == nil {
			resp.Body.Close()
		}
	}
}

func testHeader(t *testing.T, srv *Server) {
	e, err := srv.Endpoint()
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	client, err := http.NewClient(context.Background(), http.WithEndpoint(e.Host))
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	reqURL := fmt.Sprintf(e.String() + "/index")
	req, err := http2.NewRequest(http2.MethodGet, reqURL, nil)
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	req.Header.Set("content-type", "application/grpc-web+json")
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	resp.Body.Close()
}

func testBind(t *testing.T, srv *Server) {
	tests := []struct {
		method string
		path   string
		code   int
	}{
		{http2.MethodPut, "/bind/path?query=query", http2.StatusOK},
		{http2.MethodPost, "/bind/path?query=query", http2.StatusOK},
	}
	e, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client, err := http.NewClient(context.Background(), http.WithEndpoint(e.Host))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, test := range tests {
		var res bindData
		reqURL := fmt.Sprintf(e.String() + test.path)
		req, err := http2.NewRequest(test.method, reqURL, bytes.NewBuffer([]byte("{\"body\": \"body\"}")))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if kratoserrors.Code(err) != test.code {
			t.Fatalf("want %v, but got %v", test, err)
		}
		if err != nil {
			continue
		}
		if resp.StatusCode != 200 {
			_ = resp.Body.Close()
			t.Fatalf("http status got %d", resp.StatusCode)
		}
		content, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("read resp error %v", err)
		}
		err = json.Unmarshal(content, &res)
		if err != nil {
			t.Fatalf("unmarshal resp error %v", err)
		}
		if res.Path != "path" || res.Query != "query" || res.Body != "body" {
			t.Errorf("expected %s, %s, %s got %s, %s, %s", "path", "query", "body", res.Path, res.Query, res.Body)
		}
	}
}

func testClient(t *testing.T, srv *Server) {
	tests := []struct {
		method string
		path   string
		code   int
	}{
		{http2.MethodGet, "/index", http2.StatusOK},
		{http2.MethodPut, "/index", http2.StatusOK},
		{http2.MethodPost, "/index", http2.StatusOK},
		{http2.MethodPatch, "/index", http2.StatusOK},
		{http2.MethodDelete, "/index", http2.StatusOK},

		{http2.MethodGet, "/index/1", http2.StatusOK},
		{http2.MethodPut, "/index/1", http2.StatusOK},
		{http2.MethodPost, "/index/1", http2.StatusOK},
		{http2.MethodPatch, "/index/1", http2.StatusOK},
		{http2.MethodDelete, "/index/1", http2.StatusOK},

		{http2.MethodGet, "/index/notfound", http2.StatusNotFound},
		{http2.MethodGet, "/errors/cause", http2.StatusInternalServerError},
		{http2.MethodGet, "/test/prefix/123111", http2.StatusOK},
	}
	e, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client, err := http.NewClient(context.Background(), http.WithEndpoint(e.Host))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, test := range tests {
		var res testData
		reqURL := fmt.Sprintf(e.String() + test.path)
		req, err := http2.NewRequest(test.method, reqURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if kratoserrors.Code(err) != test.code {
			t.Fatalf("want %v, but got %v", test, err)
		}
		if err != nil {
			continue
		}
		if resp.StatusCode != 200 {
			_ = resp.Body.Close()
			t.Fatalf("http status got %d", resp.StatusCode)
		}
		content, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("read resp error %v", err)
		}
		err = json.Unmarshal(content, &res)
		if err != nil {
			t.Fatalf("unmarshal resp error %v", err)
		}
		if res.Path != test.path {
			t.Errorf("expected %s got %s", test.path, res.Path)
		}
	}
	for _, test := range tests {
		var res testData
		err := client.Invoke(context.Background(), test.method, test.path, nil, &res)
		if kratoserrors.Code(err) != test.code {
			t.Fatalf("want %v, but got %v", test, err)
		}
		if err != nil {
			continue
		}
		if res.Path != test.path {
			t.Errorf("expected %s got %s", test.path, res.Path)
		}
	}
}
