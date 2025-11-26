//go:build generate
// +build generate

package staffv1

//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative staff.proto