.PHONY: genswag genproto run gomigrate migrateup migratedown goupdate

run:
	go run cmd/app/main.go

genswag:
	protoc -I . --openapiv2_out ./docs --openapiv2_opt allow_merge=true,disable_default_errors=true $(file)

genproto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/v1/*.proto

gomigrate:
	migrate create -ext sql -dir db/migrations -seq $(file)

migrateup:
	migrate --path db/migrations --database 'postgresql://root:password@localhost:5432/task?sslmode=disable' -verbose up

migratedown:
	migrate --path db/migrations --database 'postgresql://root:password@localhost:5432/task?sslmode=disable' -verbose down

goupdate:
	go get -t -u ./...
	go mod tidy