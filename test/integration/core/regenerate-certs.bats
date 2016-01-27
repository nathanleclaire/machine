#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

use_shared_machine

@test "$DRIVER: regenerate the certs" {
  # Temporary skip if driver is dind.  It doesn't like the part in the code
  # where we try to rename /etc/hosts (the sed call for setting 127.0.1.1
  # hostname for Debian-based OSes) 
  #
  # TODO (nathanleclaire): regenerate-certs shouldn't be doing that anyway,
  # ideally.  Fix it and remove the skip.
  skip_if_driver dind
  run machine regenerate-certs -f $NAME
  [[ ${status} -eq 0 ]]
}

@test "$DRIVER: make sure docker still works" {
  run docker $(machine config $NAME) version
  [[ ${status} -eq 0 ]]
}
