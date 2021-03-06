set -ex

cd "`dirname $0`"

INSTALL_DIR="`pwd`/_install"
BUILD_DIR="`pwd`/_build"
PROTO_DIR="`pwd`/proto"

mkdir -p $BUILD_DIR/test/cpp
mkdir -p test/unity/JsonifUnityTest/Assets/Generated

$INSTALL_DIR/protoc/bin/protoc.exe -I$PROTO_DIR --go_out=. $PROTO_DIR/extensions.proto

go build -o $BUILD_DIR/test/protoc-gen-jsonif-cpp.exe cmd/protoc-gen-jsonif-cpp/main.go
go build -o $BUILD_DIR/test/protoc-gen-jsonif-unity.exe cmd/protoc-gen-jsonif-unity/main.go

pushd test/proto
  $INSTALL_DIR/protoc/bin/protoc.exe \
    -I. \
    -I$PROTO_DIR \
    --plugin=protoc-gen-jsonif-cpp=$BUILD_DIR/test/protoc-gen-jsonif-cpp.exe \
    --jsonif-cpp_out=$BUILD_DIR/test/cpp \
    bytes.proto \
    empty.proto \
    enumpb.proto \
    importing.proto \
    message.proto \
    nested.proto \
    oneof.proto \
    repeated.proto \
    jsonfield.proto \
    optimistic.proto \
    discard_if_default.proto \
    no_serializer.proto
  $INSTALL_DIR/protoc/bin/protoc.exe \
    --plugin=protoc-gen-jsonif-unity=$BUILD_DIR/test/protoc-gen-jsonif-unity.exe \
    --jsonif-unity_out=../unity/JsonifUnityTest/Assets/Generated \
    empty.proto \
    enumpb.proto \
    importing.proto \
    message.proto \
    nested.proto \
    oneof.proto \
    repeated.proto
popd

g++ test/cpp/main.cpp -I $BUILD_DIR/test/cpp -I $INSTALL_DIR/boost/include/ -o $BUILD_DIR/test/cpp/test.exe
$BUILD_DIR/test/cpp/test.exe