#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "$DRIVER: create with supported engine options" {
  run machine create -d $DRIVER \
    --engine-label spam=eggs \
    --engine-opt dns=8.8.8.8 \
    --engine-insecure-registry registry.myco.com \
    $NAME
  echo "$output"
  [ $status -eq 0 ]
}

@test "$DRIVER: check for engine label" {
  spamlabel=$(docker $(machine config $NAME) info | grep spam)
  [[ $spamlabel =~ "spam=eggs" ]]
}

@test "$DRIVER: check for engine option --dns" {
  [ $(docker $(machine config $NAME) run busybox nslookup google.com | grep "8.8.8.8" | wc -l) -ne 0 ]
}
