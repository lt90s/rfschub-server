package rfschub_server

//go:generate protoc --proto_path=repository/proto --micro_out=repository/proto --go_out=repository/proto repository/proto/repository.proto
//go:generate protoc --proto_path=gits/proto --micro_out=gits/proto --go_out=gits/proto gits/proto/gits.proto
//go:generate protoc --proto_path=account/proto --micro_out=account/proto --go_out=account/proto account/proto/account.proto
//go:generate protoc --proto_path=project/proto --micro_out=project/proto --go_out=project/proto project/proto/project.proto
//go:generate protoc --proto_path=index/proto --micro_out=index/proto --go_out=index/proto index/proto/index.proto
//go:generate protoc --proto_path=syntect/proto --micro_out=syntect/proto --go_out=syntect/proto syntect/proto/syntect.proto
