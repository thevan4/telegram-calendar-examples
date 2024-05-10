.PHONY: gen
gen:
	curl -O https://raw.githubusercontent.com/thevan4/telegram-calendar-examples/main/standalone_service/proto/telegram_calendar.proto && \
	mv telegram_calendar.proto vendor_protogen/telegram-calendar-examples/standalone_service && \
	for file in ./vendor_protogen/telegram-calendar-examples/standalone_service/*.proto; do \
    		base_out=$$(grep 'option go_package' $$file | sed -e 's/option go_package = "//' -e 's/;.*//'); \
    		out_dir="./$$(echo $$base_out | cut -d '/' -f 5-)"; \
    		mkdir -p $$out_dir; \
    		protoc -I./vendor_protogen \
    		       --proto_path=./vendor_protogen/telegram-calendar-examples/standalone_service \
    		       --go_out=$$out_dir \
    		       --go_opt=paths=source_relative \
    		       --go-grpc_out=$$out_dir \
    		       --go-grpc_opt=paths=source_relative \
    		       --grpc-gateway_out=$$out_dir \
    		       --grpc-gateway_opt=paths=source_relative,logtostderr=true,generate_unbound_methods=true \
    		       $$file; \
    	done
