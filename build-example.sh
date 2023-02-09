#!/bin/bash

set -xeo pipefail

#copy dockerfile into the root of the directory
cp examples/firefox/Dockerfile .

#build and run docker
docker build -t wanjohiryan/warp:firefox-x264 .

#on build remove the Dockerfile, in case we create an 'example'
rm Dockerfile 

set -e
#run the image for 4 minutes
docker run wanjohiryan/warp:firefox-x264 &

sleep 20
kill %?docker

exit 0