#!/bin/sh
docker run --name goplay-connector --rm -i -t -p 9934:9934 --link goplay-master jennal/goplay-connector --master-host goplay-master