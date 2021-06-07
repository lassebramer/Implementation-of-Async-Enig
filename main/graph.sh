#!/bin/bash

for f in ./out/*.dot; do
  dot $f -Tpng -o $f.png;
done

