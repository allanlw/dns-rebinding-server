#!/bin/bash

# go files are formatted using go fmt
go fmt ./dnsrebinder/...

# html files are formatted using html-beautify
find . -name \*.html -print0 | xargs -0 html-beautify -r -s 2

# js files are formatted using standard
find . -name \*.js -print0 | xargs -0 standard --fix
