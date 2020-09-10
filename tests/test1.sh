#!/bin/bash

cd $(dirname $0)/..
DIR=${DIR:=fakery}
[[ -d $DIR ]] || mkdir -p "$DIR"
rm -rf "$DIR"/*

DEBUG=${DEBUG:-false}

CONF=${CONF:-tests/configs/includes-regular/nginx.conf}
echo "editing config: $CONF"

./crossplane-go --debug=${DEBUG} edit $CONF tests/change1.json "$DIR"
