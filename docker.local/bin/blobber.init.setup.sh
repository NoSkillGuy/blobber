#!/bin/sh

# we are creating necessary directories for n blobber's
n=6
for i in $(seq 1 $n)
do
  mkdir -p docker.local/blobber"$i"/files
  mkdir -p docker.local/blobber"$i"/data/postgresql
  mkdir -p docker.local/blobber"$i"/log
done
