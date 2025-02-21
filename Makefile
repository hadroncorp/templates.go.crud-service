add-migration:
	goose -env .goose.env -s create ${NAME} sql

up-migration:
	goose -env .goose.env up

down-migration:
	goose -env .goose.env down
