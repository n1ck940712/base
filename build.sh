#!/bin/bash
BASE_DIR=$(dirname "$0")
DEFAULT_REGISTRY_PATH="docker-registry.r4espt.com/r4pid/"
DOCKER_DIR="$BASE_DIR/build/package"
VALID_IMAGES=("api" "websocket" "gameloop-lol-tower" "gameloop-lol-couple" "gameloop-fifa-shootup" "main")
VALID_BUILDS=("dev" "release" "ephemeral")

help() {
    echo "USAGE: build.sh [OPTIONS]"
    echo ""
    echo "Build minigame go images [${VALID_IMAGES[@]}]"
    echo ""
    echo "Options:"
    echo "    -i    Optional: Images to build. Valid images [${VALID_IMAGES[@]}]. if not set builds all" #to remove
    echo "    -b    Optional: Image type to build [${VALID_BUILDS[@]}]. Defaults: ${VALID_BUILDS[0]}"
    echo "    -v    Required: Version of the image. This will be used as the tag postfix."
    echo "    -r    Optional: Registry path. ex my.registry/name/"
    echo ""
    echo "Samples: ./build.sh -i ${VALID_IMAGES[0]} -b ${VALID_BUILDS[1]} -v "3.7-slim-buster" -r "$DEFAULT_REGISTRY_PATH""
    echo "         ./build.sh ${VALID_IMAGES[@]} -b ${VALID_BUILDS[1]} -v "3.7-slim-buster" -r "$DEFAULT_REGISTRY_PATH""
    exit 1
}

#params: 1-message
show_error() {
    echo "Error: $1"
    echo ""
    help
}

IMAGES=()
IMAGE_TYPE=${VALID_BUILDS[0]}
VERSION_TAG=
REGISTRY_PATH=$DEFAULT_REGISTRY_PATH
TAR_CONTAINER="IMAGES"
for ARG in $@; do
    if [[ "$ARG" == "-i" || "$ARG" == "--image" ]]; then
        IMAGES=
        TAR_CONTAINER="IMAGES"
    elif [[ "$ARG" == "-b" || "$ARG" == "--build" ]]; then
        IMAGE_TYPE=
        TAR_CONTAINER="BUILD"
    elif [[ "$ARG" == "-v" || "$ARG" == "--version" ]]; then
        TAR_CONTAINER="VERSION"
    elif [[ "$ARG" == "-r" || "$ARG" == "--registry" ]]; then
        REGISTRY_PATH=
        TAR_CONTAINER="REGISTRY"
    elif [[ "$ARG" == "-h" || "$ARG" == "--help" ]]; then
        help
    elif [[ $TAR_CONTAINER == "IMAGES" ]]; then
        if echo ${VALID_IMAGES[@]} | grep -q -w "$ARG"; then
            if [ -z "$IMAGES" ]; then
                IMAGES=()
            fi
            IMAGES+=("$ARG")
        else
            show_error "Invalid image $ARG"
        fi
    elif [[ $TAR_CONTAINER == "BUILD" ]]; then
        if echo ${VALID_BUILDS[@]} | grep -q -w "$ARG"; then
            IMAGE_TYPE="$ARG"
        else
            show_error "Invalid build $ARG"
        fi
    elif [[ $TAR_CONTAINER == "VERSION" ]]; then
        VERSION_TAG="$ARG"
    elif [[ $TAR_CONTAINER == "REGISTRY" ]]; then
        REGISTRY_PATH="$ARG"
    elif [[ "$ARG" =~ "-" ]]; then
        show_error "Illegal option $ARG"
    fi
done

#if image is empty set to default images
if [[ ${#IMAGES[@]} == 0 ]]; then
    IMAGES=(${VALID_IMAGES[@]})
fi

#error check 2
if [ -z "$IMAGES" ]; then
    show_error "Image is required."
elif [ -z "$IMAGE_TYPE" ]; then
    show_error "Build is required."
elif [ -z "$VERSION_TAG" ]; then
    show_error "Version is required."
elif [ -z "$REGISTRY_PATH" ]; then
    show_error "Registry path is required."
fi

# DATETIME=$(date '+%Y%m%d%H%M')
for ((i = 0; i < ${#IMAGES[@]}; ++i)); do
    IMAGE=${IMAGES[$i]}
    BASE_TAG="$REGISTRY_PATH""minigame-backend-golang-""$IMAGE"
    TAGW_BUILD="$BASE_TAG:$IMAGE_TYPE"
    TAGW_VERSION="$TAGW_BUILD-$VERSION_TAG"
    DOCKERFILE="$DOCKER_DIR/Dockerfile.$IMAGE"

    echo "BUILDING $DOCKERFILE $IMAGE_TYPE"
    if [ "$IMAGE_TYPE" == "release" ]; then
        TAG_LATEST="$BASE_TAG:latest"
        echo "docker build . -f $DOCKERFILE -t "$TAGW_BUILD" -t "$TAGW_VERSION" -t "$TAG_LATEST""
        docker build . -f $DOCKERFILE -t "$TAGW_BUILD" -t "$TAGW_VERSION" -t "$TAG_LATEST"
    else
        echo "docker build . -f $DOCKERFILE -t "$TAGW_BUILD" -t "$TAGW_VERSION""
        docker build . -f $DOCKERFILE -t "$TAGW_BUILD" -t "$TAGW_VERSION"
    fi

    if [[ $i == 0 ]]; then
        echo "$TAGW_VERSION" >.last_built_image
    else
        echo "$TAGW_VERSION" >>.last_built_image
    fi
done
