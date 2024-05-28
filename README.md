# kratos 框架拓展

> 对[kratos](https://github.com/go-kratos/kratos)的transport层进行扩展
> 
> PS: 纯粹玩玩, 缝合怪

# 拓展框架

## http
- [x] [fiber](https://github.com/gofiber/fiber)
- [x] [hertz](https://github.com/cloudwego/hertz)
- [ ] [echo](https://github.com/labstack/echo)
- [ ] [gin](https://github.com/gin-gonic/gin)

## rpc

- [ ] [arpc](https://github.com/lesismal/arpc)
- [ ] [rpcx](https://github.com/smallnest/rpcx)

# 用法

## http

### fiber

#### 代码生成

```shell
# 安装protobuf代码生成
go install google.golang.org/protobuf/cmd/protoc-gen-go
# 按住kratos代码生成
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2
# 安装fiber代码生成
go install github.com/LiangQinghai/kratos-ext/cmd/protoc-gen-go-fiber
# 生成
protoc --proto_path=. \
        --proto_path=./third_party \
        --go_out=paths=source_relative:./ \
        --go-http_out=paths=source_relative:./ \
        --go-grpc_out=paths=source_relative:./ \
        --go-fiber_out=paths=source_relative:./ \
        xxx.proto
```

### hertz

#### 代码生成

```shell
# 安装protobuf代码生成
go install google.golang.org/protobuf/cmd/protoc-gen-go
# 按住kratos代码生成
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2
# 安装hertz代码生成
go install github.com/LiangQinghai/kratos-ext/cmd/protoc-gen-go-hertz
# 生成
protoc --proto_path=. \
        --proto_path=./third_party \
        --go_out=paths=source_relative:./ \
        --go-http_out=paths=source_relative:./ \
        --go-grpc_out=paths=source_relative:./ \
        --go-hertz_out=paths=source_relative:./ \
        xxx.proto
```