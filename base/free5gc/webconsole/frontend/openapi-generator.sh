#! /bin/bash

openapi-generator-cli generate \
                      -i webconsole.yaml \
                      -g typescript-axios \
                      -o src/api

sed 's/: Time/: Date/g' src/api/api.ts >src/api/api.ts.mod
mv src/api/api.ts.mod src/api/api.ts
