#!/bin/bash

set -e

# Check if $USERNAME exists
if [ -z "$USERNAME" ]; then

    # Set $USERNAME to $USER
    USERNAME=$USER
fi

####################################################
#####  Global varaiables
####################################################

CWD=`pwd`

REGISTRY=registry.internal.tensor.works
REPO="sps-"$USERNAME
VERSION="0.0.0-devel"

PROFILER_IMAGE=pod-profiler-gatherer
PROFILER_DOCKERFILE=.dockerfiles/pod-profiler-gatherer.dockerfile

BUILD_PROFILER=false

BUILD_ACTION_COMPILE_GOLANG=true
BUILD_ACTION_CONTAINER_BUILD=true
BUILD_ACTION_CONTAINER_PUSH=true
BUILD_ACTION_DELETE_POD=false

# Defines whether or not the golang source has been compiled for this script execution.
# This flag is used to prevent golang compiling for multiple container builds
COMPILE_GOLANG_HAS_RUN=false

####################################################
#####  Flags
####################################################

while test $# -gt 0; do
  case "$1" in

    # custom version
    --version)
        shift
        VERSION=$1
        shift
    ;;
    --repo)
        shift
        REPO=$1
        shift
    ;;
    --registry)
        shift
        REGISTRY=$1
        shift
    ;;
    --skip-compile)
        BUILD_ACTION_COMPILE_GOLANG=false
        shift
    ;;
    --skip-build)
        BUILD_ACTION_CONTAINER_BUILD=false
        shift
    ;;
    --skip-push)
        BUILD_ACTION_CONTAINER_PUSH=false
        shift
    ;;
    --profiler)
        BUILD_PROFILER=true
        shift
    ;;
    --dockerhub)
        REGISTRY="docker.io"
        REPO="tensorworks"
        shift
    ;;
    --pod-del)
        BUILD_ACTION_DELETE_POD=true
        shift
    ;;
    --all)
        BUILD_PROFILER=true
        shift
    ;;
    *)
        break
        ;;
  esac
done

####################################################
#####  Computed variables
####################################################

PROFILER_TAG=$REGISTRY/$REPO/$PROFILER_IMAGE:$VERSION

####################################################
#####  Variable dump
####################################################

echo "=====  Build variables  ====="

echo "------ Globals-------"
echo "REGISTRY:               $REGISTRY"
echo "REPO:                   $REPO"
echo "VERSION:                $VERSION"
echo ""

echo "----- Container -----"
echo "PROFILER_IMAGE:         $PROFILER_IMAGE"
echo ""

echo "------ Building ------"
echo "PROFILER:                 $BUILD_PROFILER"
echo ""

echo "------ Actions ------"
echo "COMPILE GOLANG:         $BUILD_ACTION_COMPILE_GOLANG"
echo "BUILD CONTAINER:        $BUILD_ACTION_CONTAINER_BUILD"
echo "PUSH  CONTAINER:        $BUILD_ACTION_CONTAINER_PUSH"
echo "DELETE POD:             $BUILD_ACTION_DELETE_POD"
echo ""
echo "-------------------------"
echo ""

echo "PROFILER TAG:             $PROFILER_TAG"
echo "======================="

####################################################
#####  Functions
####################################################

# This will Build a container
#   $1 - Container Tag
#   $2 - Docker File
#   $3 - Docker build additional arguments
#   $4 - Docker build additional arguments
#   $5 - Docker build additional arguments
function build_container(){
    TAG=$1
    DOCKER_FILE=$2

    if [ "$BUILD_ACTION_CONTAINER_BUILD" == "true" ]; then
        echo ""
        echo "Building Container " $TAG $DOCKER_FILE
        echo ""
        docker build --tag $TAG --file $DOCKER_FILE $3 $4 $5 $6 $7 $8 .
    fi
}

# This will push the container to the registry
#   $1 - Container Tag
function push_container(){
    TAG=$1

    if [ "$BUILD_ACTION_CONTAINER_PUSH" == "true" ]; then
        docker push $TAG
    fi
}

# Generates and compiles the golang source code
function compile_golang(){
    # Check if we're compiling the golang source
    if [ "$BUILD_ACTION_COMPILE_GOLANG" == "true" ] && [ "$COMPILE_GOLANG_HAS_RUN" != "true" ]; then

        echo "Building Go source..."
        go run build.go --generate --release
        COMPILE_GOLANG_HAS_RUN=true
    fi
}

# Builds and pushes the container. Uses dockerignore file and takes a string that lists all files to omit from the ignore
#   $1 - Container tag
#   $2 - Dockerfile
#   $3 - A list of files to omit from dockerignore as a relative path from the repo root (e.g. '\n!bin/linux/amd64/pod-profiler-gatherer\n!bin/linux/amd64/pod-profiler-gatherer')
function build_push(){
    TAG=$1
    DOCKERFILE=$2
    OMIT_IGNORE=$3

    # Create a docker ignore file
    echo -e "*!$OMIT_IGNORE" > .dockerignore

    # Build the container
    build_container $TAG $DOCKERFILE

    # Clean up the docker ignore file
    rm -rf .dockerignore

    # Push the container
    push_container $TAG
}

####################################################
#####  Main
####################################################

# Build the profiler
if [ "$BUILD_PROFILER" == "true" ]; then
    compile_golang
    build_push $PROFILER_TAG $PROFILER_DOCKERFILE "\n!bin/linux/amd64/pod-profiler-gatherer"
fi


if [ "$BUILD_ACTION_DELETE_POD" == true ]; then
   
    POD_PREFIX_NAMES=()
    
    if [ "$BUILD_PROFILER" == true ] ; then 
        POD_PREFIX_NAMES+=($PROFILER_IMAGE)
    fi 

    echo ""
    echo "Getting Pod to delete"

    for POD_PREFIX_NAME in ${POD_PREFIX_NAMES[@]}; do

        echo $POD_PREFIX_NAME
        echo "Deleteing pod with prefix ${POD_PREFIX_NAME}"

        POD_NAME=`kubectl get pods --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep ${POD_PREFIX_NAME}`
        kubectl delete pod $POD_NAME

        echo ""
    done
fi
