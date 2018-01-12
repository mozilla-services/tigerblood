
.PHONY: loadtest build-db start-db setup-db rm-db run

loadtest:
	HAWK_ID=root HAWK_KEY=toor locust --host=http://localhost:8000 -f tools/loadtesting/locustfile.py

build-db:
	docker build -f postgres.Dockerfile -t postgres-ip4r .

start-db:
	# create a postgres container bound to port 5432 locally
	docker run --detach --name postgres-ip4r -p 127.0.0.1:5432:5432 postgres-ip4r

setup-db:
	# create tigerblood user and database
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres -c "CREATE ROLE tigerblood WITH LOGIN;"'
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres -c "CREATE DATABASE tigerblood;"'
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE tigerblood TO tigerblood;"'
	docker exec -ti postgres-ip4r bash -c 'psql -U postgres tigerblood -c "CREATE EXTENSION ip4r;"'

rm-db:
	docker rm -f postgres-ip4r

test:
	TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" go test

coverage:
	TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" go test -coverprofile=coverage.txt -covermode=atomic
	sed "s|_$$(pwd)/|./|g" coverage.txt > rel-coverage.txt
	go tool cover -html=rel-coverage.txt

build:
	go build ./cmd/tigerblood/

build-static:
	CGO_ENABLED=0 go build --ldflags '-extldflags "-static"' ./cmd/tigerblood/

run:
	TIGERBLOOD_BIND_ADDR=127.0.0.1:8080 \
		TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" \
		TIGERBLOOD_HAWK=true \
		TIGERBLOOD_PROFILE=true \
		TIGERBLOOD_DATABASE_MAX_OPEN_CONNS=5 \
		TIGERBLOOD_DATABASE_MAX_IDLE_CONNS=5 \
		TIGERBLOOD_DATABASE_MAXLIFETIME=24h \
			./tigerblood --config-file config.yml
