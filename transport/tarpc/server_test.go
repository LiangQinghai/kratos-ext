package tarpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/lesismal/arpc"
	"github.com/lesismal/arpc/log"
	"net"
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
		var in string
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
				//fmt.Printf("server recieved: %v\n", req)
				return "Hello server.——From server", nil
			},
		)
		reply, _ := handler(ctx, in)
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
	client, err := arpc.NewClient(
		func() (net.Conn, error) {
			return net.DialTimeout("tcp", srv.endpoint.Host, time.Second*4)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()
	req := &MessageWrapper{
		Data: []byte("Hello server.——From client"),
		Headers: map[string][]string{
			"Content-Type": {"text/plain"},
		},
	}
	var rsp MessageWrapper
	err = client.Call("/echo", &req, &rsp, time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("client read: [data: %v, header: %v, err: %v] \n", string(rsp.Data), rsp.Headers, rsp.Err)
	time.Sleep(time.Second)
	if srv.Stop(ctx) != nil {
		t.Errorf("expected nil got %v", srv.Stop(ctx))
	}
}

func TestNewServer(t *testing.T) {
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
	clients := make([]*arpc.Client, clientNum)
	for i := 0; i < clientNum; i++ {
		client, err := arpc.NewClient(
			func() (net.Conn, error) {
				return net.DialTimeout("tcp", srv.endpoint.Host, time.Second*4)
			},
		)
		if err != nil {
			t.Fatal("NewClient failed:", err)
			return
		}
		clients[i] = client
		defer client.Stop()
	}

	for i := 0; i < clientNum; i++ {
		client := clients[i]
		for j := 0; j < eachClientCoroutineNum; j++ {
			go func() {
				var err error
				var data = make([]byte, 512)
				for k := 0; true; k++ {
					_, _ = rand.Read(data)
					req := &MessageWrapper{
						Data: []byte("Hello server.——From client"),
						Headers: map[string][]string{
							"Content-Type": {"text/plain"},
						},
					}
					rsp := &MessageWrapper{}
					err = client.Call("/echo", req, rsp, time.Second*5)
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
