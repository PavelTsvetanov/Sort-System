# sort-system
[![Build Status](https://travis-ci.com/PavelTsvetanov/sort-system.svg?branch=main)](https://travis-ci.com/PavelTsvetanov/sort-system)

This project is a version of the [SORT system](https://www.youtube.com/watch?v=BQDliV7w7_8), built during the course "Go to Production" by Ocado

## How to run the project
* `make grpc-compile` to generate all grpc-related files in the `gen/` folder
* Enter `sorting-service` and type `make go-run`
* Enter `fulfilment-service` and type `make go-run`
* (Optional) Run scripts/seed-orders.sh to test basic scenario