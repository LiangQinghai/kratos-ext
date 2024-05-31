package tarpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	grpc2 "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/lesismal/arpc/log"
	"google.golang.org/grpc"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var headerMid = func(handler middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		if tr, ok := transport.FromServerContext(ctx); ok {
			tr.ReplyHeader().Set("Content-Type", "text/plain")
		}
		return handler(ctx, req)
	}
}

var helloWorldEcho = func(srv *Server) HandlerFunc {
	return func(c *Ctx) {
		ctx, bytes, err := srv.DecodeRequest(c)
		if err != nil {
			panic(err)
		}
		var in TestReq
		err = srv.DecodeData(bytes, &in)
		if err != nil {
			panic(err)
		}
		SetOperation(ctx, "op")
		handler := srv.Middleware(
			ctx,
			func(ctx context.Context, req interface{}) (interface{}, error) {
				//if tr, ok := transport.FromServerContext(ctx); ok {
				//	ct := tr.RequestHeader().Get("Content-Type")
				//	fmt.Printf("server recieved ct header: %s\n", ct)
				//}
				//fmt.Printf("server recieved: %v\n", req.(*TestReq).Message)
				return &TestReply{
					Message: "Hello server.——From server",
				}, nil
			},
		)
		reply, _ := handler(ctx, &in)
		// invoke method
		resp := srv.EncodeResponse(
			ctx,
			reply,
			nil,
		)
		srv.Write(c, resp)
	}
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(Middleware(
		headerMid,
	))
	srv.Handle("/echo", helloWorldEcho(srv))
	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}
	go func() {
		if err := srv.Start(ctx); err != nil {
			//panic(err)
		}
	}()
	client, err := Dail(
		context.Background(),
		WithEndpoint(srv.endpoint.Host),
	)
	if err != nil {
		t.Fatal(err)
	}
	req := TestReq{
		Message: "Hello server.——From client",
	}
	var rsp TestReply
	err = client.Call(context.Background(), "/echo", &req, &rsp)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("client read: [data: %v, err: %v] \n", rsp.Message, err)
	time.Sleep(time.Second)
	if srv.Stop(ctx) != nil {
		t.Errorf("expected nil got %v", srv.Stop(ctx))
	}
}

func TestNewServer(t *testing.T) {
	//bGrpc(t)
	//fmt.Println("----------------------------------------")
	//fmt.Println("----------------------------------------")
	//fmt.Println("----------------------------------------")
	//fmt.Println("----------------------------------------")
	//fmt.Println("----------------------------------------")
	bArpc(t)
}

type testService struct {
	UnsafeTestGrpcServiceServer
}

func (t *testService) Echo(_ context.Context, _ *TestReq) (*TestReply, error) {
	return &TestReply{
		Message: "hello client",
	}, nil
}

func bGrpc(t *testing.T) {
	srv := grpc2.NewServer(grpc2.Middleware(
		metadata.Server(),
	))
	timeout := 1 * time.Minute
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	RegisterTestGrpcServiceServer(srv, new(testService))
	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}
	go func() {
		if err := srv.Start(ctx); err != nil {
			//panic(err)
		}
	}()

	var (
		qpsSec                 uint64
		qpsTotal               uint64
		clientNum              = runtime.NumCPU() * 2
		eachClientCoroutineNum = 10
	)
	clients := make([]TestGrpcServiceClient, clientNum)
	for i := 0; i < clientNum; i++ {
		endpoint, err := srv.Endpoint()
		client, err := grpc2.DialInsecure(
			context.Background(),
			grpc2.WithOptions(grpc.WithIdleTimeout(0)),
			grpc2.WithEndpoint(endpoint.Host),
			grpc2.WithTimeout(30*time.Second),
			grpc2.WithMiddleware(
				metadata.Client(),
			),
		)
		if err != nil {
			t.Fatal("NewClient failed:", err)
			return
		}
		serviceClient := NewTestGrpcServiceClient(client)
		clients[i] = serviceClient
		defer func(client *grpc.ClientConn) {
			err := client.Close()
			if err != nil {

			}
		}(client)
	}

	for i := 0; i < clientNum; i++ {
		client := clients[i]
		for j := 0; j < eachClientCoroutineNum; j++ {
			go func() {
				var err error
				var data = make([]byte, 512)
				for k := 0; true; k++ {
					_, _ = rand.Read(data)
					_, _ = client.Echo(context.Background(), &TestReq{Message: "hhh"})
					if err != nil {
						t.Errorf("Call failed: %v", err)
					} else {
						atomic.AddUint64(&qpsSec, 1)
					}
				}
			}()
		}
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		for i := 0; true; i++ {
			if _, ok := <-ticker.C; !ok {
				return
			}
			if i < 3 {
				fmt.Printf("[qps preheating %v: %v]", i+1, atomic.SwapUint64(&qpsSec, 0))
				continue
			}

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			var u uint64 = 1024 * 1024
			fmt.Printf("----------------------------------\n")
			fmt.Printf("TotalAlloc: %d\n", mem.TotalAlloc/u)
			fmt.Printf("Alloc: %d\n", mem.Alloc/u)
			fmt.Printf("HeapAlloc: %d\n", mem.HeapAlloc/u)
			fmt.Printf("HeapSys: %d\n", mem.HeapSys/u)
			qps := atomic.SwapUint64(&qpsSec, 0)
			qpsTotal += qps
			fmt.Printf(
				"[qps: %v], [avg: %v / s], [total: %v, %v s]\n",
				qps,
				int64(float64(qpsTotal)/float64(i-2)),
				qpsTotal,
				int64(float64(i-2)),
			)
			fmt.Printf("----------------------------------\n")
		}
	}()
	select {
	case <-ctx.Done():
		fmt.Printf("done\n")
	case <-time.After(timeout):
		fmt.Printf("timeout\n")
	}
	if srv.Stop(ctx) != nil {
		t.Errorf("expected nil got %v", srv.Stop(ctx))
	}

}

func bArpc(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()
	log.SetLevel(log.LevelNone)
	timeout := 1 * time.Minute
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	srv := NewServer(Middleware(
		headerMid,
	))
	srv.Handle("/echo", helloWorldEcho(srv))
	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}
	go func() {
		if err := srv.Start(ctx); err != nil {
			//panic(err)
		}
	}()

	var (
		qpsSec                 uint64
		qpsTotal               uint64
		clientNum              = runtime.NumCPU() * 2
		eachClientCoroutineNum = 10
	)
	clients := make([]*Client, clientNum)
	for i := 0; i < clientNum; i++ {
		client, err := Dail(
			context.Background(),
			WithEndpoint(srv.endpoint.Host),
		)
		if err != nil {
			t.Fatal("NewClient failed:", err)
			return
		}
		clients[i] = client
	}

	for i := 0; i < clientNum; i++ {
		client := clients[i]
		for j := 0; j < eachClientCoroutineNum; j++ {
			go func() {
				var err error
				var data = make([]byte, 512)
				for k := 0; true; k++ {
					_, _ = rand.Read(data)
					req := TestReq{Message: "Hello server.——From client"}
					var rsp TestReply
					err = client.Call(context.Background(), "/echo", &req, &rsp)
					if err == nil {
						atomic.AddUint64(&qpsSec, 1)
					}
				}
			}()
		}
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		for i := 0; true; i++ {
			if _, ok := <-ticker.C; !ok {
				return
			}
			if i < 3 {
				fmt.Printf("[qps preheating %v: %v]", i+1, atomic.SwapUint64(&qpsSec, 0))
				continue
			}

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			var u uint64 = 1024 * 1024
			fmt.Printf("----------------------------------\n")
			fmt.Printf("TotalAlloc: %d\n", mem.TotalAlloc/u)
			fmt.Printf("Alloc: %d\n", mem.Alloc/u)
			fmt.Printf("HeapAlloc: %d\n", mem.HeapAlloc/u)
			fmt.Printf("HeapSys: %d\n", mem.HeapSys/u)
			qps := atomic.SwapUint64(&qpsSec, 0)
			qpsTotal += qps
			fmt.Printf(
				"[qps: %v], [avg: %v / s], [total: %v, %v s]\n",
				qps,
				int64(float64(qpsTotal)/float64(i-2)),
				qpsTotal,
				int64(float64(i-2)),
			)
			fmt.Printf("----------------------------------\n")
		}
	}()
	select {
	case <-ctx.Done():
		fmt.Printf("done\n")
	case <-time.After(timeout):
		fmt.Printf("timeout\n")
	}
	if srv.Stop(context.Background()) != nil {
		t.Errorf("expected nil got %v", srv.Stop(ctx))
	}
}
