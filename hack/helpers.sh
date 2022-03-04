#!/bin/bash

#	_push_to_git_server pushes a given git repository to a git server
#	deployed in kubernetes.
#
#		- name: 	name of the gitserver obj
#		- namespace:	name of the ns where the git server exists
#		- url:		url of the git repository to clone & push
push_to_private_git_server() {
        local name=$1
        local namespace=$2
        local url=$3

        local app_scratch=$(mktemp -d)
        local attempts=0
        local port
        local proxy_pid

        kubectl -n $namespace rollout status deployment $name --watch

        git -C $app_scratch clone $url .
        port=$(_available_port)

        (
                kubectl -n $namespace port-forward service/$name $port:80 &
                trap "kill $! || true" EXIT

                git -C $app_scratch \
                        remote set-url \
                        origin http://localhost:$port/$(basename $url)

                until git -C $app_scratch push origin HEAD; do
                        test $attempts -eq 10 && {
                                echo "aborting. failed to push."
                                exit 1
                        }

                        ((attempts = attempts + 1))

                        sleep 1
                done
        )
}
