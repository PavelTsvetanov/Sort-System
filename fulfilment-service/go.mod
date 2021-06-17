module github.com/PavelTsvetanov/sort-system/fulfilment-service

go 1.16

replace github.com/PavelTsvetanov/sort-system/gen => ../gen

require (
	github.com/PavelTsvetanov/sort-system/gen v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.38.0
)
