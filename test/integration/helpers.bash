#!/bin/bash

function echo_to_log {
    echo "$BATS_TEST_NAME
----------
$output
----------

"   >> ${BATS_LOG}
}

function teardown {
    echo_to_log
}

function errecho {
    >&2 echo "$@"
}

function only_if_env {
    if [[ ${!1} != "$2" ]]; then
        errecho "This test requires the $1 environment variable to be set to $2. Skipping..."
        skip
    fi
}

function require_env {
    if [[ -z ${!1} ]]; then
        errecho "This test requires the $1 environment variable to be set in order to run."
        exit 1
    fi
}

function skip_if_driver () {
    for SKIPPED_DRIVER in "$@"; do
        if [[ "$DRIVER" == "$SKIPPED_DRIVER" ]]; then
            skip
        fi
    done
}

# Make sure these aren't set while tests run (can cause confusing behavior)
unset DOCKER_HOST DOCKER_TLS_VERIFY DOCKER_CERT_DIR
