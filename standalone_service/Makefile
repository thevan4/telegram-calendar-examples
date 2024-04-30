.PHONY: gen
gen:
	for file in ./proto/*.proto; do \
		base_out=$$(grep 'option go_package' $$file | sed -e 's/option go_package = "//' -e 's/;.*//'); \
		out_dir="./$$(echo $$base_out | cut -d '/' -f 5-)"; \
		mkdir -p $$out_dir; \
		protoc -I./vendor_protogen \
		       --proto_path=./proto \
		       --go_out=$$out_dir \
		       --go_opt=paths=source_relative \
		       --go-grpc_out=$$out_dir \
		       --go-grpc_opt=paths=source_relative \
		       --grpc-gateway_out=$$out_dir \
		       --grpc-gateway_opt=paths=source_relative,logtostderr=true,generate_unbound_methods=true \
		       --openapiv2_out=$$out_dir \
		       --openapiv2_opt=logtostderr=true \
		       $$file; \
	done
