#!/bin/bash

for val in {1..9}
do
    echo VM$val Updating
    ssh $1@fa22-cs425-220$val.cs.illinois.edu "cd ./mp2-hangy6-tian23; git pull origin $2; exit"
    # ssh hangy6@fa22-cs425-220$val.cs.illinois.edu "cd ./mp1-hangy6-tian23; git checkout hangy6; git fetch --all; git reset --hard origin/main; git pull; exit"
    echo VM$val Updated
done
echo VM10 Updating
ssh $1@fa22-cs425-2210.cs.illinois.edu "cd ./mp2-hangy6-tian23; git pull origin $2; exit"
# ssh hangy6@fa22-cs425-2210.cs.illinois.edu "cd ./mp1-hangy6-tian23; git checkout hangy6; git fetch --all; git reset --hard origin/main; git pull; exit"
echo VM10 Updated

echo "All VMs Have Been Updated!"