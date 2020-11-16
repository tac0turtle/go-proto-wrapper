install:
	@echo "Installing gowrapper."
	go install ./protoc-gen-gowrapper

gen-examples:
	@echo "Generating Examples"
	buf generate --path ./examples
