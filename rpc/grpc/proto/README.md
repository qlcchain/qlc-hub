# How to generate files
```bash

# grpc apis
protoc -I. -I$GOPATH/src -I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.7/third_party/googleapis --go_out=plugins=grpc:. types.proto

# grpc-gateway apis 
protoc -I. -I$GOPATH/src -I$I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.7/third_party/googleapis --grpc-gateway_out=logtostderr=true:. types.proto

# swagger apis
protoc -I. -I$GOPATH/src -I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.7/third_party/googleapis --swagger_out=logtostderr=true:. types.proto

```
