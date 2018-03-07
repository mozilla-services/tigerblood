
.PHONY: loadtest build-db start-db setup-db rm-db run

.env:
	cp .env.example .env

loadtest:
	HAWK_ID=root HAWK_KEY=toor locust --host=http://localhost:8000 -f tools/loadtesting/locustfile.py

test:
	TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" go test

test-container: .env
	docker-compose run test test

coverage:
	TIGERBLOOD_DSN="user=tigerblood dbname=tigerblood sslmode=disable" go test -coverprofile=coverage.txt -covermode=atomic
	sed "s|_$$(pwd)/|./|g" coverage.txt > rel-coverage.txt
	go tool cover -html=rel-coverage.txt

build:
	go build ./cmd/tigerblood/

build-cli:
	go build ./cmd/tigerblood-cli/

build-container: .env
	docker-compose build

clean-cli:
	rm -f ./tigerblood-cli

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

run-container: .env
	docker-compose run web web --config-file config.yml
