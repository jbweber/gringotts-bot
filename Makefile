shell = bash

.PHONY: generate-mocks
generate-mocks:
	go install github.com/vektra/mockery/v2@v2.33.1
	mockery
