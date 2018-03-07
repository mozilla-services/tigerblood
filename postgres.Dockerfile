FROM postgres:9.5

RUN apt-get update \
 && apt-get install -qy postgresql-9.5-ip4r \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

# CircleCI does not support mounting folders, so add it directly
# https://circleci.com/docs/2.0/building-docker-images/#mounting-folders
ADD bin/db/init_db.sh /docker-entrypoint-initdb.d
