#!/bin/bash

cd "$(dirname "$0")"

cd demo_files

python -m SimpleHTTPServer 1337
