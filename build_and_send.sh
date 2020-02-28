#!/bin/bash

env GOOS=linux GOARCH=amd64 go build
scp bst_api bst@35.196.119.150:/home/bst/bst_api

ssh bst@35.196.119.150