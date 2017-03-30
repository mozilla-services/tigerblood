
.PHONY: loadtest start-db setup-db rm-db

loadtest:
	HAWK_ID=root HAWK_KEY=toor locust --host=http://localhost:8000 -f tools/loadtesting/locustfile.py

start-db:
	docker run --detach --name postgres-ip4r -p 127.0.0.1:5432:5432 postgres-ip4r

setup-db:
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres -c "CREATE ROLE tigerblood WITH LOGIN;"'
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres -c "CREATE DATABASE tigerblood;"'
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE tigerblood TO tigerblood;"'
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres tigerblood -c "CREATE EXTENSION ip4r;"'

rm-db:
	docker rm -f postgres-ip4r

test:
	TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" go test

torch-cpu:
	go-torch --seconds 30 http://127.0.0.1:8080/debug/pprof/profile && open torch.svg

mem-objs:
	go tool pprof --alloc_objects http://127.0.0.1:8080/debug/pprof/heap

build-gc-opts:  # log gc optimizations like inlined funcs
	rm -f gc-build.out
	go build -v -gcflags=-m &> gc-build.out

coverage:
	TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" go test -coverprofile=coverage.txt -covermode=atomic
	sed "s|_$$(pwd)/|./|g" coverage.txt > rel-coverage.txt
	go tool cover -html=rel-coverage.txt

build:
	go build ./cmd/tigerblood/

build-static:
	CGO_ENABLED=0 go build --ldflags '-extldflags "-static"' ./cmd/tigerblood/
