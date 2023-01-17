#!/bin/bash

set -o pipefail

echo "Has the drive been mounted to /mnt?(y/n): "
read $FLAG
if [[ $FLAG != "y" ]]; then
    echo "Drive is mounted, proceeding..."
else
    echo "Mount the drive to proceed."
    exit 1
fi

if [[ -d /home/kayleigh/git/endless_waltz ]]; then
    echo "Found repos in path, off we go..."
    cp -r /home/kayleigh/git/endless_waltz /tmp
    cp -r /home/kayleigh/git/ew-rtl-entropy /tmp
else
    echo "Could not find repos in expected path!"
    exit 2
fi

cd /tmp
echo "Removing files for backup..."
rm -v $(find . -exec file {} \; | grep -i elf | cut -d ':' -f 1)
rm -rfv endless_waltz/.git

echo "Creating tarballs..."
file1=$(date +%s)-endless_waltz.tar.gz 
tar -vczf $file1 endless_waltz
echo "Tarred endless_waltz!"

rm -rfv ew-rtl-entropy/.git
file2=$(date +%s)-ew-rtl-entropy.tar.gz 
tar -vczf $file2 ew-rtl-entropy
echo "Tarred ew-rtl-entropy!"

files=($file1 $file2)
for file in "$files"; do
    if [[ $(ls -alh $file  | cut -d ' ' -f 5 | sed 's/K//') -gt 900 ]]; then 
	echo "WARNING: EW Tarball over 9000!"
	echo "Exiting gracefully instead of copying..."
	exit 3
    fi
done

echo 
echo
echo "**************************"
echo "ABOUT TO SUDO, NEED INPUT:"
echo "**************************"
sudo cp /tmp/$file1 /mnt
sudo cp /tmp/$file2 /mnt

echo "File copy complete!"
echo "Good on you for running a backup <3"
