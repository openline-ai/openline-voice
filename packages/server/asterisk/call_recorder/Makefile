generate: proto/messagestore.proto
	 python3 -m grpc_tools.protoc -I . --python_out=. --pyi_out=. --grpc_python_out=.  proto/messagestore.proto

proto/messagestore.proto:
	mkdir -p proto
	curl https://raw.githubusercontent.com/openline-ai/openline-customer-os/otter/packages/server/message-store/proto/messagestore.proto -o proto/messagestore.proto

clean:
	rm -rf proto
