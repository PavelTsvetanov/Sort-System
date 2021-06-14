module github.com/PavelTsvetanov/sort-system/sorting-service

go 1.16

replace github.com/PavelTsvetanov/sort-system/gen => ../gen

require (
	github.com/PavelTsvetanov/sort-system/gen v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.37.1 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)
