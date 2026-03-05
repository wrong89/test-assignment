include .env
export

service-run:
	go run main.go

migrate-up:
	migrate -path migrations -database ${DB_URL} up

migrate-down:
	migrate -path migrations -database ${DB_URL} down