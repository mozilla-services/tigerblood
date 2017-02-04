build:
	chmod +x decay.py
	pip install psycopg2 -t .
	zip -r tigerblood-decay.zip decay.py .
build-static:
	chmod +x decay.py
	tar xvzf psycopg2-static.tgz
	zip -r tigerblood-decay.zip decay.py .
clean:
	rm -rf tigerblood-decay.zip *.dist-info *.pyc psycopg2/
test:
	DECAY_RATE=1 DB_DSN="host=127.0.0.1 user=tigerblood dbname=tigerblood sslmode=disable" python decay.py
.PHONY: clean build test build-static
