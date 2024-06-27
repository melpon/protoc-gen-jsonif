set -ex

cd "`dirname $0`"

INSTALL_DIR="`pwd`/_install"
BUILD_DIR="`pwd`/_build"
PROTO_DIR="`pwd`/proto"

rm -rf $BUILD_DIR/test/cpp
rm -rf $BUILD_DIR/test/c
rm -rf $BUILD_DIR/test/typescript
rm -rf test/unity/JsonifUnityTest/Assets/Generated
mkdir -p $BUILD_DIR/test/cpp
mkdir -p $BUILD_DIR/test/c
mkdir -p $BUILD_DIR/test/typescript
mkdir -p test/unity/JsonifUnityTest/Assets/Generated

$INSTALL_DIR/protoc/bin/protoc -I$PROTO_DIR --go_out=. $PROTO_DIR/extensions.proto

go build -o $BUILD_DIR/test/protoc-gen-jsonif-cpp cmd/protoc-gen-jsonif-cpp/main.go
go build -o $BUILD_DIR/test/protoc-gen-jsonif-c cmd/protoc-gen-jsonif-c/main.go
go build -o $BUILD_DIR/test/protoc-gen-jsonif-unity cmd/protoc-gen-jsonif-unity/main.go
go build -o $BUILD_DIR/test/protoc-gen-jsonif-typescript cmd/protoc-gen-jsonif-typescript/main.go

pushd test/proto
  $INSTALL_DIR/protoc/bin/protoc \
    -I. \
    -I$PROTO_DIR \
    --plugin=protoc-gen-jsonif-cpp=$BUILD_DIR/test/protoc-gen-jsonif-cpp \
    --jsonif-cpp_out=$BUILD_DIR/test/cpp \
    bytes.proto \
    empty.proto \
    enumpb.proto \
    importing.proto \
    message.proto \
    nested.proto \
    oneof.proto \
    repeated.proto \
    size.proto \
    jsonfield.proto \
    optimistic.proto \
    optional.proto \
    discard_if_default.proto \
    no_serializer.proto
  $INSTALL_DIR/protoc/bin/protoc \
    -I. \
    -I$PROTO_DIR \
    --plugin=protoc-gen-jsonif-c=$BUILD_DIR/test/protoc-gen-jsonif-c \
    --jsonif-c_out=$BUILD_DIR/test/c \
    bytes.proto \
    empty.proto \
    enumpb.proto \
    importing.proto \
    message.proto \
    nested.proto \
    oneof.proto \
    optional.proto \
    repeated.proto \
    size.proto
  $INSTALL_DIR/protoc/bin/protoc \
    --plugin=protoc-gen-jsonif-unity=$BUILD_DIR/test/protoc-gen-jsonif-unity \
    --jsonif-unity_out=../unity/JsonifUnityTest/Assets/Generated \
    empty.proto \
    enumpb.proto \
    importing.proto \
    message.proto \
    nested.proto \
    oneof.proto \
    optional.proto \
    repeated.proto
  $INSTALL_DIR/protoc/bin/protoc \
    --plugin=protoc-gen-jsonif-typescript=$BUILD_DIR/test/protoc-gen-jsonif-typescript \
    --jsonif-typescript_out=$BUILD_DIR/test/typescript \
    empty.proto \
    enumpb.proto \
    importing.proto \
    message.proto \
    nested.proto \
    oneof.proto \
    optional.proto \
    repeated.proto
popd

g++ test/cpp/main.cpp \
  -I $BUILD_DIR/test/cpp \
  -I $INSTALL_DIR/boost/include/ \
  -o $BUILD_DIR/test/cpp/test
$BUILD_DIR/test/cpp/test

g++ test/cpp/main.cpp \
  -I $BUILD_DIR/test/cpp \
  -I $INSTALL_DIR/json/include/ \
  -o $BUILD_DIR/test/cpp/test_nlohmann \
  -DJSONIF_USE_NLOHMANN_JSON
$BUILD_DIR/test/cpp/test_nlohmann

g++ -g \
  test/c/main.cpp \
  $BUILD_DIR/test/c/*.cpp \
  $BUILD_DIR/test/c/google/protobuf/*.cpp \
  -I $BUILD_DIR/test/c \
  -I $BUILD_DIR/test/cpp \
  -I $INSTALL_DIR/boost/include/ \
  -o $BUILD_DIR/test/c/test
$BUILD_DIR/test/c/test

# 複数ファイルで同じ変数や関数を定義している場合にリンクエラーにならないことを確認する
gcc \
  test/c/multidef_1.c \
  test/c/multidef_2.c \
  test/c/multidef.cpp \
  $BUILD_DIR/test/c/*.cpp \
  $BUILD_DIR/test/c/google/protobuf/*.cpp \
  -lstdc++ \
  -I $BUILD_DIR/test/c \
  -I $BUILD_DIR/test/cpp \
  -I $INSTALL_DIR/boost/include/ \
  -o $BUILD_DIR/test/c/test_multidef
$BUILD_DIR/test/c/test

# TypeScript のテスト
pushd test/typescript
  rm -rf dist/
  npx webpack
  node dist/main.js
popd
