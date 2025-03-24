#!/usr/bin/env bash

# Copyright (c) F5, Inc.
# This source code is licensed under the Apache License, Version 2.0 license found in the
# LICENSE file in the root directory of this source tree.

# shellcheck shell=bash
# Generate support file from a public git repo
# Only call it through //go:generate comments in project root to update .gen.go files

set -e

branch="master"
url=""
sub_path=""
genArgs=()

help() {
    cat << EOF
Generate support file from a public repo(currently only git).
Only call me through //go:generate comments in project root to update .gen.go files.
If the command line is long, we recommend using json config(--config-path).
If you don't provide --path, we will use the whole repository at provided url by default.

usage: $(basename "$0") [-b|--branch] [-c|--config-path] [-d|--directive-map-name]
 [-mn|--match-func-name] [-f|--filter] [-o | --override] [-mc|--match-func-comment] [-p|--path] [--url] [-h|--help]
    -h  | --help                Display this message

    -b  | --branch              Branch to checkout, defaults to "$branch". (optional)

    -c  | --config-path         The path of json config file. Normally the json config file should be under nginx-go-crossplane/scripts/generate/configs.
     The json config will be unmarshed into generator.GenerateConfig(at nginx-go-crossplane/internal/generator/generator.go).
     The file can contain directiveMapName, matchFuncName, matchFuncComment, filter, and override.
     They provide same functions as other arguments directive-map-name, match-func-name, match-func-comment, filter, and override. (optional)

    -d  | --directive-map-name  Name of the generated directive map. You should provide it here or through json config(--config-path).
     Normally it should start with lowercase to avoid export.
     If this is provided, the directive_map_name in json config will be ignored.

    -mn | --match-func-name     Name of the generated matchFunc. You should provide it here or through json config(--config-path).
     Normally it should start with uppercase to export.
     If this is provided, the match_func_name in json config will be ignored.

    -f  | --filter              The directives you want to exclude from output. An example is: -filter directive1 -filter directive2.
     You can provide it here or through json config(--config-path). If this is provided, the filter in json config will be ignored. (optional)
    
    -o  | --override            Strings used to override the output. It should follow the format:{directive:bitmask00|bitmask01...,bitmask10|bitmask11...}.
     An example is --override log_format:ngxHTTPMainConf|ngxConf2More,ngxStreamMainConf|ngxConf2More.
     To use | and , in command line, you may need to enclose your input in quotes, i.e. --override 'directive:mask1,mask2,...'.
     You can provide it here or through json config(--config-path). You can provide it multiple times for different directives.
     If this is provided, the override in json config will be ignored. (optional)
    
    -mc | --match-func-comment  The code comment for generated matchFunc.
	 You can add some explanations like which modules included in it. Normally it should start with match-func-name.
	 If this is provided, the matchFuncComment in json config will be ignored. (optional)
    
    -p  | --path                Path to a directory in the repository containing the source code of the nginx module. (optional)
    
    --url                       Url used for git clone. (required)
EOF
}

while [ ! $# -eq 0 ]; do
    case "$1" in
        --help | -h)
            help
            exit 0
            ;;
        --branch | -b)
            branch="$2"
            shift
            ;;
        --config-path | -c)
            genArgs+=("--config-path=$2")
            shift
            ;;
        --directive-map-name | -d)
            genArgs+=("--directive-map-name=$2")
            shift
            ;;
        --match-func-name | -mn)
            genArgs+=("--match-func-name=$2")
            shift
            ;;
        --filter | -f)
            genArgs+=("-filter=$2")
            shift
            ;;
        -override | -o)
            genArgs+=("-override=\"$2\"")
            shift
            ;;
        --match-func-comment | --mc)
            genArgs+=("--match-func-comment=\"$2\"")
            shift
            ;;
        --path | -p)
            sub_path="$2"
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


tmp=$(mktemp -d)
cleanup() {
    rm -rf "$tmp"
}
trap cleanup EXIT

genArgs+=("--src-path=$tmp")

if [ "$sub_path" = "" ]; then
    git clone "$url" "$tmp" --depth 1 --branch "$branch" -q 1>&2
else
    (
        git clone "$url" "$tmp" --no-checkout --depth 1 --branch "$branch" -q 1>&2 &&
        cd "$tmp" &&
        # sparse-checkout doesn't support -q option
        git sparse-checkout init --cone 1>&2  &&
        git sparse-checkout set "$sub_path" 1>&2 &&
        git checkout -q 1>&2
    )
fi

# --match-fun-comment may contain spaces, to avoid auto quotes we go this way.
sh -c "go run ./cmd/generate ${genArgs[*]}"
