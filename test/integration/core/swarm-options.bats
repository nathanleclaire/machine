#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash
export TOKEN=$(curl -sS -X POST "https://discovery-stage.hub.docker.com/v1/clusters")

export QUEEN_NAME="$NAME-queenbee"
export WORKER_NAME="$NAME-workerbee"

@test "create swarm master" {
    run machine create -d $DRIVER --swarm --swarm-master --swarm-discovery "token://$TOKEN" --swarm-strategy binpack --swarm-opt heartbeat=5 $QUEEN_NAME
    echo ${output}
    [[ "$status" -eq 0 ]]
}

@test "create swarm node" {
    run machine create -d $DRIVER --swarm --swarm-discovery "token://$TOKEN" $WORKER_NAME
    [[ "$status" -eq 0 ]]
}

@test "ensure strategy is correct" {
    strategy=$(docker $(machine config --swarm $QUEEN_NAME) info | grep "Strategy:" | awk '{ print $2 }')
    echo ${strategy}
    [[ "$strategy" == "binpack" ]]
}

@test "ensure heartbeat" {
    heartbeat_arg=$(docker $(machine config $QUEEN_NAME) inspect -f '{{index .Args 9}}' swarm-agent-master)
    echo ${heartbeat_arg}
    [[ "$heartbeat_arg" == "--heartbeat=5" ]]
}
