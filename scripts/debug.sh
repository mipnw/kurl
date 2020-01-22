#!/bin/bash

BINARY="/usr/local/bin/kurl"
GO_PACKAGE="kurl/src/kurl_cmd"

set -e
me=`basename $0`

function print_help() {
    echo "$me [options] -- [your binary arguments]"
    echo "launch $BINARY in the delve headless debugger on port 2345"
    echo 
    echo "options:"
    echo "  --source     build $GO_PACKAGE from sources and launch headless debugger"
    echo "               useful for rapid iterations with source code synchronization between localhost and docker container"
    echo "  --nodlv      don't attach delve, just run the server"
    echo 
    echo "Notes"
    echo "If CTRL-C is unable to kill headless delve and you haven't yet attached a client debugger to, try attaching a client debugger"
}

source=
debug=1
while [[ $# > 0 ]]; do
    case $1 in 
        --help|help )
            print_help
            exit 0
            ;;
        --source )
            source=1
            shift 1
            ;;
        --nodlv )
            debug=0
            shift 1
            ;;
        -- )
            # whatever appears beginning with -- we pass to the target
            break
            ;;
        * )
            print_help
            exit 1
            ;;
    esac
done

# whatever appears beginning with -- we pass to the target
args=$@

if [[ $debug == 1 ]]; then
    if [[ -z $source ]]; then
        echo
        echo "Running under headless debugger"
        set -x
        dlv exec $BINARY --headless --listen=:2345 --log --api-version 2 $args
        { set +x; } 2>/dev/null
    else
        echo
        echo "Building from source and running under headless debugger"
        set -x
        dlv debug $GO_PACKAGE --headless --listen=:2345 --log --api-version 2 --build-flags "--mod vendor" $args
        { set +x; } 2>/dev/null
    fi
else
    if [[ -z $source ]]; then
        set -x
        $BINARY $args
        { set +x; } 2>/dev/null
    else
        set -x
        go run -mod vendor $GO_PACKAGE $args
        { set +x; } 2>/dev/null
    fi
fi

