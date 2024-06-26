#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "unreal_nibi_sdk.h"
#if defined(_WIN64)
#include <io.h>
#else
#include <unistd.h>
#endif
int main()
{
    // Switch to the test network
    int ret = NewNibiruClientDefault();

    if (ret != 0) {
        fprintf(stderr, "Failed to connect testnet1\n");
        return 1;
    }

    // Define key names and mnemonics
    char *keyNameAdmin = "AdminKey";
    char *keyName = "TestKey";
    char *adminMnemonic = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host";
    char *testMnemonic = "toe cream coach quiz cactus nest spike gauge opinion legal father stadium lizard match wood immune odor depart sauce timber crash pig thought seat";
    char *passPhrase = "";

    // Create accounts
    if (CreateAccount(keyNameAdmin, adminMnemonic, passPhrase) != 0 ||
        CreateAccount(keyName, testMnemonic, passPhrase) != 0)
    {
        fprintf(stderr, "Failed to create accounts\n");
        return 1;
    }

    // Print account addresses
    printf("AdminKey Address: %s\n", GetAddressFromKeyName(keyNameAdmin));
    printf("TestKey Address: %s\n", GetAddressFromKeyName(keyName));

    char *adminAddress = GetAddressFromKeyName(keyNameAdmin);
    char *testAddress = GetAddressFromKeyName(keyName);

    char *denom = "unibi";

    // Transfer token from AdminKey to TestKey
    printf("Transferring tokens from AdminKey to TestKey...\n");
    char *testTx = TransferToken(testAddress, adminAddress, denom, 2500);
    if (testTx == NULL)
    {
        fprintf(stderr, "Failed to transfer tokens\n");
        return 1;
    }
    printf("Transfer successful. Transaction hash: %s\n", testTx);
    sleep(10);

    // Execute Wasm contract
    printf("Executing Wasm contract...\n");
    // Payload for minting a new NFT
    char *payload = "{\"mint\": {\"token_id\": \"unique-nft-18\", \"owner\": \"nibi1zy7amen6h5e4whcta4ac656l0whsalzmnqrkc5\", \"token_uri\": \"https://metadata.com/nft1.json\"}}";
    // Address of the deployed contract
    char *contractAddress = "nibi1xs48fjdmq0u5rg6txhrrc5n7xlxstew6pvm82hsh6ftplyuysdaqkdkzfk";
    char *testTx2 = ExecuteWasmContract(testAddress, contractAddress, payload, denom, 1);
    if (testTx2 == NULL)
    {
        fprintf(stderr, "Failed to execute Wasm contract\n");
        return 1;
    }
    printf("Execution successful. Transaction hash: %s\n", testTx2);
    sleep(10);
    QueryTXHash(testTx2);

    // Query Wasm contract for NFT ownership
    printf("Querying Wasm contract for NFT ownership...\n");
    char *query = "{\"owner_of\": {\"token_id\": \"unique-nft-18\", \"include_expired\": false}}";
    char *responseMsg = QueryWasmContract(contractAddress, query);
    if (responseMsg == NULL)
    {
        fprintf(stderr, "Failed to query Wasm contract\n");
        return 1;
    }
    printf("Query successful. Response: %s\n", responseMsg);

    return 0;
}
