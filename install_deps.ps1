$ErrorActionPreference = 'Stop'

$BUILD_DIR = Join-Path (Resolve-Path ".").Path "_build"
$INSTALL_DIR = Join-Path (Resolve-Path ".").Path "_install"

mkdir $BUILD_DIR -Force
mkdir $INSTALL_DIR -Force

# protoc
$PROTOC_VERSION = "22.3"

$_URL = "https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-win64.zip"
$_FILE = "protoc-$PROTOC_VERSION-win64.zip"
# ダウンロードと展開
Push-Location $BUILD_DIR
  if (!(Test-Path $_FILE)) {
    Invoke-WebRequest -Uri $_URL -OutFile $_FILE
  }
  if (Test-Path "protoc") {
    Remove-Item protoc -Force -Recurse
  }
  if (Test-Path "${INSTALL_DIR}\protoc") {
    Remove-Item ${INSTALL_DIR}\protoc -Force -Recurse
  }
  # Expand-Archive -Path $_FILE -DestinationPath .
  mkdir protoc -Force
  Push-Location protoc
    7z x ..\$_FILE
  Pop-Location
  Move-Item ".\protoc\" "${INSTALL_DIR}\protoc\"
Pop-Location

# Boost
$BOOST_VERSION = "1.82.0"

$_BOOST_UNDERSCORE_VERSION = $BOOST_VERSION.Replace(".", "_")
$_URL = "https://boostorg.jfrog.io/artifactory/main/release/${BOOST_VERSION}/source/boost_${_BOOST_UNDERSCORE_VERSION}.zip"
$_FILE = "boost_${_BOOST_UNDERSCORE_VERSION}.zip"
# ダウンロードと展開
Push-Location $BUILD_DIR
  if (!(Test-Path $_FILE)) {
    Invoke-WebRequest -Uri $_URL -OutFile $_FILE
  }
  if (Test-Path "boost_${_BOOST_UNDERSCORE_VERSION}") {
    Remove-Item boost_${_BOOST_UNDERSCORE_VERSION} -Force -Recurse
  }
  # Expand-Archive -Path $_FILE -DestinationPath .
  7z x $_FILE
Pop-Location

# インストール
Push-Location $BUILD_DIR\boost_${_BOOST_UNDERSCORE_VERSION}
  & .\bootstrap.bat
  & .\b2.exe headers
  if (Test-Path "$INSTALL_DIR\boost") {
    Remove-Item $INSTALL_DIR\boost -Force -Recurse
  }
  mkdir -Force $INSTALL_DIR\boost\include
  Copy-Item -Recurse boost $INSTALL_DIR\boost\include\
Pop-Location

# nlohmann::json
Push-Location $INSTALL_DIR
  if (Test-Path "json") {
    Remove-Item json -Force -Recurse
  }
  git clone https://github.com/nlohmann/json.git
Pop-Location