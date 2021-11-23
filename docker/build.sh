#!/bin/bash

if [ "$1" = "no-cache" ]; then
    docker buildx build --no-cache --platform linux/amd64 -t kuaifan/sdwan:manage-0.0.1 --push -f ./manage.Dockerfile . &
    docker buildx build --no-cache --platform linux/amd64 -t kuaifan/sdwan:nginx-0.0.1 --push -f ./nginx.Dockerfile . &
    docker buildx build --no-cache --platform linux/amd64 -t kuaifan/sdwan:work-0.0.1 --push -f ./work.Dockerfile . &
else
    docker buildx build --platform linux/amd64 -t kuaifan/sdwan:manage-0.0.1 --push -f ./manage.Dockerfile . &
    docker buildx build --platform linux/amd64 -t kuaifan/sdwan:nginx-0.0.1 --push -f ./nginx.Dockerfile . &
    docker buildx build --platform linux/amd64 -t kuaifan/sdwan:work-0.0.1 --push -f ./work.Dockerfile . &
fi

wait