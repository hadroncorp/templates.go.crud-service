add-migration:
	goose -env .goose.env -s create ${NAME} sql

up-migration:
	goose -env .goose.env up

down-migration:
	goose -env .goose.env down

# This follows the convention of the package name according to big tech companies style guides.
gen-mock:
	mockgen -source=${PACKAGE}/${FILE}.go -destination=${PACKAGE}mock/${FILE}.go -package=${PACKAGE}mock

run-all-tests:
	go test ./... -tags=integration,e2e -cover -race -p=10
