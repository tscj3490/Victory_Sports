#!/bin/bash
echo "checking that the workdir and uploads dir exists"
if [ ! -d ./victory-frontend_workdir/uploads ]; then
  mkdir -p ./victory-frontend_workdir/uploads;
fi
