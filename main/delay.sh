#!/bin/bash

mkdir -p out

date >> out/delay.log

# delay network
for i in {1000..10000..1000}
do
  echo "started delay/final $i"
  echo "started delay/final $i" >> out/delay.log
  go run setup.go 1800 10 900 1 2 $i &>> out/delay.log
  echo "finished delay/final $i"
  sleep 5
  echo "started delay/no-final $i"
  echo "started delay/no-final $i" >> out/delay.log
  go run setup.go 1800 10 900 0 2 $i &>> out/delay.log
  echo "finished delay/no-final $i"
  sleep 5
done
