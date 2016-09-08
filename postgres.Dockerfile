FROM postgres:9.5

RUN apt-get update \
 && apt-get install -qy postgresql-9.5-ip4r \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*
