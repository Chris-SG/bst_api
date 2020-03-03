#!/bin/bash

env GOOS=linux GOARCH=amd64 go build

ssh -t bst@35.196.119.150 "sudo systemctl stop bst_api"
scp bst_api bst@35.196.119.150:/home/bst/bst_api
ssh -t bst@35.196.119.150 "sudo systemctl start bst_api"