#! /bin/bash

# prerequisites
#  - docker

# use Docker to run OpenAPI Generator
docker run --rm -v $PWD:/local openapitools/openapi-generator-cli generate -i /local/webconsole.yaml -g typescript-axios -o /local/src/api

# replace Time with Date in the file
sed 's/: Time/: Date/g' /local/src/api/api.ts > /local/src/api/api.ts.mod
mv /local/src/api/api.ts.mod /local/src/api/api.ts # rename the replaced file to the original file