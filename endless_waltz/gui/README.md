implementing build container and script for github actions builds

#!/bin/bash
#cd gui
cp -r . /tmp/foo
for i in `find . -maxdepth 1 -type l -ls | tr -s ' ' | cut -d ' ' -f 12 | cut -d '/' -f 2`; do cp --remove-destination ../common/$i /tmp/foo/; done

