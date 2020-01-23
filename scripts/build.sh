#!/bin/bash

GOMODULE="github.com/mipnw/kurl"

LOCAL_PATH="/usr/local/bin"
LOCAL_IMAGES=("kurl")
LOCAL_GO_PACKAGES=("kurl_cmd")

CROSSCOMPILE_PATH="bin"
CROSSCOMPILE_IMAGES=("kurl")
CROSSCOMPILE_GO_PACKAGES=("kurl_cmd")

set -e
me=`basename $0`

function print_help() {
    echo "$me [flags][options]"
    echo "Build the image for this container, and optionally cross compile it for your localhost"
    echo
    echo "Flags (required):"
    echo "  --debug   produce a debug build with symbols"
    echo "  --release produce a release build without symbols"
    echo
    echo "Options:"
    echo "  --binplace place the output binary at /usr/local/bin"
    echo "  --mac     produce a client compiled for mac in addition to the regular build"
    echo "  --win     produce a client compiled for windows in addition to the regular build"
    echo "  --linux   produce a client compile for linux in addition to the regular build"
    echo 
}

os=
debug=
binplace=
while [[ $# > 0 ]]; do
    case $1 in 
        --help|help )
            print_help
            exit 0
            ;;
        --debug )
            debug=1
            shift 1
            ;;
        --release )
            debug=0
            shift 1
            ;;
        --binplace )
            binplace=1
            shift 1
            ;;
        --mac )
            os=darwin
            shift 1
            ;;
        --win )
            os=windows
            shift 1
            ;;
        --linux )
            os=linux
            shift 1
            ;;
        * )
            print_help
            exit 1
            ;;
    esac
done


echo
if [[ -z $debug ]]; then
    print_help
    exit 1
fi

if [[ -z $output && -n $os ]]; then
    echo "when using the [--mac|--linux|--windows] flag, you must also use the --binplace flag"
    echo "otherwise we're trying to write both binaries at the same location"
    exit 1
fi

# We either binplace, or output at $GOPATH/bin
output="-o bin/${LOCAL_IMAGES[$i]}"
[[ -n $binplace ]] && output="-o $LOCAL_PATH/${LOCAL_IMAGES[$i]}"

for i in ${!LOCAL_IMAGES[@]}; do
    if [[ $debug == "1" ]]; then
        # build the debug build
        echo "Building debug $LOCAL_PATH/${LOCAL_IMAGES[$i]}"
        set -x
        env CGO_ENABLED=0 \
        go build \
            $output \
            -mod vendor \
            -gcflags "all=-N -l" \
            "$GOMODULE/${LOCAL_GO_PACKAGES[$i]}"
        { set +x; } 2>/dev/null
    else 
        # build the release build
        echo "Building release $LOCAL_PATH/${LOCAL_IMAGES[$i]}"
        set -x
        env CGO_ENABLED=0 \
        go build \
            $output \
            -mod vendor \
            -ldflags="-w -s" \
            "$GOMODULE/${LOCAL_GO_PACKAGES[$i]}"
        { set +x; } 2>/dev/null
    fi
done

arch=amd64
if [[ -n $os ]]; then
    for i in ${!CROSSCOMPILE_IMAGES[@]}; do
        echo
        echo "Building $CROSSCOMPILE_PATH/${CROSSCOMPILE_IMAGES[$i]} for your workstation: os=$os/$arch"
        set -x
        env GOOS=$os GOARCH=$arch \
        go build \
            -o "$CROSSCOMPILE_PATH/${CROSSCOMPILE_IMAGES[$i]}" \
            -mod vendor \
            -ldflags="-w -s" \
            "$GOMODULE/${CROSSCOMPILE_GO_PACKAGES[$i]}"
        { set +x; } 2>/dev/null
    done
fi
