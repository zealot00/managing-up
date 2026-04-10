module github.com/zealot/example-mcp-sdk

go 1.25.0

require (
	github.com/mark3labs/mcp-go v0.47.1
	github.com/zealot/managing-up/sdk v0.0.0
)

require (
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
)

replace github.com/zealot/managing-up/sdk => ../../../sdk
