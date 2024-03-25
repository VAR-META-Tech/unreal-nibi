#include <stdio.h>
#include <stdlib.h>
#include "unreal_nibi_sdk.h"

int main() {
    // Create a new NibiruClient instance using the exported Go function.
    int ret = NewNibiruClientDefault();
    if (ret != 0) {
        printf("Failed to create NibiruClient\n");
        return 1;
    }
    //Create new wallet
    char* prases = GenerateRecoveryPhrase();
    printf("Prases %s", prases);
    // Generate Menomonic
    // Create key(private,public =>signner) from menemonic
    // Storage in keyring
    // Get wallet address
    return 0;
}