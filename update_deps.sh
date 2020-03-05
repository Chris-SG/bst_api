#!/bin/bash

sha=$(git ls-remote git://github.com/chris-sg/bst_server_models.git HEAD | awk '{ print $1}')
go get github.com/chris-sg/bst_server_models@$sha

sha=$(git ls-remote git://github.com/chris-sg/eagate.git HEAD | awk '{ print $1}')
go get github.com/chris-sg/eagate@$sha

sha=$(git ls-remote git://github.com/chris-sg/eagate_db.git HEAD | awk '{ print $1}')
go get github.com/chris-sg/eagate_db@$sha

sha=$(git ls-remote git://github.com/chris-sg/eagate_models.git HEAD | awk '{ print $1}')
go get github.com/chris-sg/eagate_models@$sha