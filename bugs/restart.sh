#! /bin/bash
ps -ef | grep -E 'go run|go-build' | grep -v grep| awk '{print $2}' | xargs kill -9
nohup ./run.sh & >> log.txt