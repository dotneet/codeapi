#!/bin/bash
set -euo pipefail

docker build -f Dockerfile.python -t python_runner .
