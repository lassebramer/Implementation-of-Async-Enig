#!/bin/bash

mkdir -p out

date >> out/partition.log

# partition network
for i in {10..100..10}
do
  echo "started partition/final $i"
  echo "started partition/final $i" >> out/partition.log
  go run setup.go 1800 10 900 1 4 $i &>> out/partition.log
  echo "finished partition/final $i"
  sleep 5
  echo "started partition/no-final $i"
  echo "started partition/no-final $i" >> out/partition.log
  go run setup.go 1800 10 900 0 4 $i &>> out/partition.log
  echo "finished partition/no-final $i"
  sleep 5
done
