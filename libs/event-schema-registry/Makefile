PROTO_FOLDERS := hadron/places/v1
GENERATED_FOLDERS := generated/go

# [COMPILERS]

all:
	$(foreach item, $(GENERATED_FOLDERS), \
		find $(item) -mindepth 1 ! \( -name 'requirements.txt' -o -name 'go.mod' -o -name 'go.sum' -o -name 'cargo.toml' -o -name 'pom.xml' \) -exec rm -rf {} +; \
	)
	$(foreach item, $(PROTO_FOLDERS), \
		$(MAKE) gen-proto-go PROTO_SRC=$(item); \
	)

gen-proto-go:
	protoc \
		--go_out=./generated/go --go_opt=module=event-schema-registry \
        ${PROTO_SRC}/*.proto