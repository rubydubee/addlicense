#!/bin/bash
#
# Script that runs the addlicense container.
# Addlicense options can be passed to this script, but make sure you wrap them
# in quotes ("").

# Examples of how to use:
# 1. Run addlicense with check option: ./run_addlicense "-check"
# 2. Run addlicense with some more options: 
#    ./run_addlicense "-v -config config.yml -c Nokia"
# 3. Run specific tag of the addlicense image: 
#    TAG="something" ./run_addlicense.sh "-check"

tag=${TAG:="latest"}
docker run -e OPTIONS="$@" --rm -it -v $(pwd):/myapp addlicense-nokia:$tag
