# unreal-nibi

Unreal client SDK for interacting with the Nibiru blockchain.

# Step to build

git submodule update --init --recursive

go build -o unreal_nibi_sdk.so -buildmode=c-shared ./api.go

gcc -o sdk_test sdk_test.c unreal_nibi_sdk.so -lpthread

./sdk_test
