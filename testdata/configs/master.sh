#!/bin/bash

# build reference payloads from python, 
# assumes crossplane is installed and in your path
cd $(dirname $0)/..
for file in $(find python -name \*.conf)
do
	into=${file%.*}.json
	echo crossplane parse $file $into
	crossplane parse $file > $into
done
