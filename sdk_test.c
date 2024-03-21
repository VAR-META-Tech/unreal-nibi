#include <stdio.h>
#include <stdlib.h>
#include "unreal_nibi_sdk.h"

int main() {
    // Create a new NibiruClientService instance using the exported Go function.
    void* clientServicePtr = NewNibiruClientService();
    if (clientServicePtr == NULL) {
        printf("Failed to create NibiruClientService\n");
        return 1;
    }

    // Get the RPC endpoint using the exported Go function.
    char* rpcEndpoint = RPCEndpoint(clientServicePtr);
    if (rpcEndpoint == NULL) {
        printf("RPC endpoint is NULL\n");
    } else {
        printf("RPC endpoint: %s\n", rpcEndpoint);
        // Remember to free the C string when done.
        free(rpcEndpoint);
    }

    // Normally, you would also need to free the clientServicePtr if it was allocated in Go.
    // However, since Go manages this memory, you should not free it in C unless you allocated
    // it in Go using C.malloc.

    return 0;
}