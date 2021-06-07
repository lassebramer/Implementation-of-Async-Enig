#!/bin/bash

mkdir -p out

date >> out/drop.log

# drop network
for i in {0..50..5}
do
  echo "started drop/final $i"
  echo "started drop/final $i" >> out/drop.log
  go run setup.go 1800 10 900 1 3 $i &>> out/drop.log
  echo "finished drop/final $i"
  sleep 5
  echo "started drop/no-final $i"
  echo "started drop/no-final $i" >> out/drop.log
  go run setup.go 1800 10 900 0 3 $i &>> out/drop.log
  echo "finished drop/no-final $i"
  sleep 5
done

