generate: proto/messagestore.proto proto/oasisapi.proto
	 python3 -m grpc_tools.protoc -I . --python_out=. --pyi_out=. --grpc_python_out=.  proto/messagestore.proto proto/oasisapi.proto

proto/messagestore.proto:
	mkdir -p proto
	curl https://raw.githubusercontent.com/openline-ai/openline-customer-os/otter/packages/server/message-store/proto/messagestore.proto -o proto/messagestore.proto

proto/oasisapi.proto:
	mkdir -p proto
	curl https://raw.githubusercontent.com/openline-ai/openline-oasis/otter/packages/server/oasis-api/proto/oasisapi.proto -o proto/oasisapi.proto

clean:
	rm -rf proto
