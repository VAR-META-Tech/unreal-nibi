# unreal-nibi

Unreal client SDK for interacting with the Nibiru blockchain.

# Step to build

git submodule update --init --recursive

go build -o unreal_nibi_sdk.so -buildmode=c-shared ./api.go

gcc -o sdk_test sdk_test.c unreal_nibi_sdk.so -lpthread

./sdk_test

# For MacOS 
go build -o unreal_nibi_sdk.dylib -buildmode=c-shared ./api.go

gcc -o sdk_test sdk_test.c unreal_nibi_sdk.dylib -lpthread

./sdk_test

# For Test Unreal UI

install_name_tool -id @rpath/unreal_nibi_sdk.dylib unreal_nibi_sdk.dylib