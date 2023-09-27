#!/bin/bash

echo "Generating candidates..."
ssh-keygen -M generate -O bits=8192 moduli-8192.candidates
echo "Done Generating candidates!"

echo "Testing candidates..."
ssh-keygen -M screen -f moduli-8192.candidates moduli-8192
echo "Candidate testing finished!"

mv moduli-8192 outfile
./moduli.py > rn.go
