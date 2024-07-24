#!/bin/bash

# Copyright (c) F5, Inc.
# This source code is licensed under the Apache License, Version 2.0 license found in the
# LICENSE file in the root directory of this source tree.

# shellcheck shell=bash
# Generate support file from a public git repo
# Only call it through //go:generate comments in project root to update .gen.go files

set -e

branch="master"
directive_map_name=""
match_func_name=""
url=""
output_path=""
sub_path=""
match_func_comment=""
help() {
    cat << EOF
Generate support file from a public repo(currently only git).
Only call me through //go:generate comments in project root to update .gen.go files.

usage: $(basename "$0") [-d|--directive-map-name] [-b|--branch] [-m|--match-func-name] [-o|--output-path]
[-c|comment-match-func] [--url] [-h|--help]
    -h  | --help                Display this message
    -d  | --directive-map-name  Name of the generated directive map. (required)
    -b  | --branch              Branch to checkout, defaults to "$branch". (optional)
    -m  | --match-func-name     Name of the generated MatchFunc. (required)
    -o  | --output-path         The path of the output file. (required)
    -s  | --sub-path            The subfolder or file within the repository contains all the directives you want, whole repository by default. (optional)
    -c  | --comment-match-func  The code comment appears above generated matchFunc named by match-func-name. No comments by default. (optional)
    --url                       Url used for git clone. (required)
EOF
}

while [ ! $# -eq 0 ]; do
    case "$1" in
        --help | -h)
            help
            exit 0
            ;;
        --directive-map-name | -d)
            directive_map_name="$2"
            shift
            ;;
        --branch | -b)
            branch="$2"
            shift
            ;;
        --match-func-name | -m)
            match_func_name="$2"
            shift
            ;;
        --output-path | -o)
            output_path="$2"
            shift
            ;;
        --sub-path | -s)
            sub_path="$2"
            shift
            ;;
        --comment-match-func | -c)
            match_func_comment="$2"
            shift
            ;;
        --url)
            url="$2"
            shift
            ;;
        *)
            printf "Invalid option: %s\n" "$1" 1>&2
            exit 1
            ;;
    esac
    shift
done

if [ "$url" = "" ]; then
    echo "url can't be empty"
    exit 1
fi

if [ "$output_path" = "" ]; then
    echo "output-path can't be empty"
    exit 1
fi

if [ "$directive_map_name" = "" ]; then
    echo "directive-map-name can't be empty"
    exit 1
fi

if [ "$match_func_name" = "" ]; then
    echo "match-func-name can't be empty"
    exit 1
fi

tmp=$(mktemp -d)
cleanup() {
    rm -rf "$tmp"
}

trap cleanup EXIT

genArgs=(
  --src-path="$tmp"
  --directive-map-name="$directive_map_name"
  --match-func-name="$match_func_name"
  --match-func-comment="$match_func_comment"
)

if [ "$sub_path" = "" ]; then
    git clone "$url" "$tmp" --depth 1 --branch "$branch"
else
    (
        git clone "$url" "$tmp" --no-checkout --depth 1 --branch "$branch"&&
        git -C "$tmp" sparse-checkout init --cone &&
        git -C "$tmp" sparse-checkout set "$sub_path" &&
        git -C "$tmp" checkout
    )
fi

go run ./cmd/generate "${genArgs[@]}" > "$output_path"