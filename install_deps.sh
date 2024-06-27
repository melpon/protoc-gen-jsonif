#!/bin/bash

cd "`dirname $0`"

INSTALL_DIR="`pwd`/_install"
BUILD_DIR="`pwd`/_build"

mkdir -p $BUILD_DIR
mkdir -p $INSTALL_DIR

go install github.com/golang/protobuf/protoc-gen-go

# protoc
PROTOC_VERSION="24.4"

_URL="https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-linux-x86_64.zip"
_FILE="protoc-$PROTOC_VERSION-linux-x86_64.zip"
pushd $BUILD_DIR
  if [ ! -e $_FILE ]; then
    curl -Lo $_FILE $_URL
  fi
  rm -rf $INSTALL_DIR/protoc
  mkdir -p $INSTALL_DIR/protoc
  unzip $_FILE -d $INSTALL_DIR/protoc
popd

# Boost
BOOST_VERSION="1.82.0"

_BOOST_UNDERSCORE_VERSION=${BOOST_VERSION//./_}
_URL="https://boostorg.jfrog.io/artifactory/main/release/${BOOST_VERSION}/source/boost_${_BOOST_UNDERSCORE_VERSION}.zip"
_FILE="boost_${_BOOST_UNDERSCORE_VERSION}.zip"
pushd $BUILD_DIR
  if [ ! -e $_FILE ]; then
    curl -Lo $_FILE $_URL
  fi
  rm -rf boost_${_BOOST_UNDERSCORE_VERSION}
  unzip $_FILE

  pushd boost_${_BOOST_UNDERSCORE_VERSION}
    ./bootstrap.sh
    ./b2 install \
      --prefix=$INSTALL_DIR/boost \
      --build-dir=$BUILD_DIR/boost-build \
      --with-filesystem \
      --with-program_options \
      --with-json \
      target-os=linux \
      address-model=64 \
      variant=release \
      link=static
  popd
popd

# nlohmann::json
pushd $INSTALL_DIR
  rm -rf json
  git clone https://github.com/nlohmann/json.git
popd
