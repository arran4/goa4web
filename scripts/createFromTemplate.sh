#!/bin/bash

if [[ $1 = "" ]]; then
  exec echo "No arg"
fi

sed "s/templatePage/${1}/g" templatePage.go > "${1}.go"
sed "s/templatePage/${1}/g" templates/templatePage.tmpl > "templates/${1}.tmpl"

git add  "${1}.go" "templates/${1}.tmpl"
