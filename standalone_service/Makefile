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
		       --openapiv2_out=./api/swagger \
		       --openapiv2_opt=allow_merge=true,merge_file_name=telegram.calendar.json \
		       $$file; \
	done


.PHONY: run
build-and-run:
	docker-compose -f deployments/build_and_run/docker-compose.yml up --build

.PHONY: run
docker-run:
	docker-compose -f deployments/docker_hub_run/docker-compose.yml up --build
