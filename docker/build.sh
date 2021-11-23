#!/bin/bash
set -e

rm -rf ./release
docker run --rm -v $(dirname "$PWD"):/sdos -w /sdos golang:buster make
/bin/cp -rf ../release ./

param="$1"
version="0.0.1"

platforms="linux/amd64"
tags="manage nginx work"
for platform in $platforms; do
    for tag in $tags; do
        if [ "$param" = "no-cache" ]; then
            docker buildx build --no-cache --platform ${platform} -t kuaifan/sdwan:${tag}-${version} --push -f ./${tag}.Dockerfile .
        else
            docker buildx build --platform ${platform} -t kuaifan/sdwan:${tag}-${version} --push -f ./${tag}.Dockerfile .
        fi
    done
done

rm -rf ./release