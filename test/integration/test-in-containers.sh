#!/bin/bash

set -e

cd ../../
./script/build -osarch="linux/amd64"
cp ./docker-machine_linux-amd64 test/integration
cd -
docker-compose kill
docker-compose rm -v -f
docker-compose build

for MCN_DRIVER in amazonec2 digitalocean; do
    for CORE_TEST in $(ls core/); do
        docker-compose run -d ${MCN_DRIVER} core/${CORE_TEST}
    done

    # The tests in cli/ run pretty fast for now, so don't bother running them
    # concurrently
    docker-compose run -d ${MCN_DRIVER} cli/

    for DRIVER_SPECIFIC_TEST in $(ls drivers/${MCN_DRIVER}); do
        docker-compose run -d ${MCN_DRIVER} ${DRIVER_SPECIFIC_TEST}
    done
done
