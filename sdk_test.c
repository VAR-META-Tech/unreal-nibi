#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "unreal_nibi_sdk.h"
#include <unistd.h>

int main()
{
    // Create a new NibiruClient instance using the exported Go function.
    // int ret = NewNibiruClientDefault();
    // if (ret != 0)
    // {
    //     printf("Failed to create NibiruClient\n");
    //     return 1;
    // }

    SwitchNetwork("test");

    char *keyNameAdmin = "AdminKey";
    char *keyName = "TestKey";
    // Create new wallet
    // Generate Menomonic
    char *testMnemonic = "toe cream coach quiz cactus nest spike gauge opinion legal father stadium lizard match wood immune odor depart sauce timber crash pig thought seat";
    char *adminMnemonic = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host";

    // Create key(private,public =>signner) from menemonic
    char *passPhrase = "";
    int createAdminAccount = CreateAccount(keyNameAdmin, adminMnemonic, "");
    if (createAdminAccount != 0)
    {
        printf("Failed to create account\n");
        return 1;
    }
    int createAccount = CreateAccount(keyName, testMnemonic, passPhrase);
    if (createAccount != 0)
    {
        printf("Failed to create account\n");
        return 1;
    }

    // Get account address
    char *address = GetAddressFromKeyName(keyName);
    char *adminAddress = GetAddressFromKeyName(keyNameAdmin);

    printf("Admin Address: %s\n", adminAddress);
    printf("Account Address: %s\n", address);

    BaseAccount *baseAccAdmin = QueryAccount(adminAddress);
    BaseAccount *baseAcc = QueryAccount(address);
    char *testTx = TransferToken(address, "nibi1dgmut4ed90ka7qze5smllk3asd6nkl3du6grwa" , "unibi", 1000000);
    if (testTx == NULL)
    {
        printf("Failed to Test transfer\n");
        return 1;
    }

    sleep(3);
    char *testTx2 = ExecuteWasmContract(adminAddress, "nibi1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3slkhcux",
                                        "{\"mint\": {\"token_id\": \"unique-nft-15\", \"owner\": \"nibi1zy7amen6h5e4whcta4ac656l0whsalzmnqrkc5\", \"token_uri\": \"https://metadata.com/nft1.json\"}}",
                                        "unibi", 1);
    if (testTx2 == NULL)
    {
        printf("Failed to Test ExecuteWasmContract\n");
        return 1;
    }
    else
    {
        printf("TxHash %s\n", testTx2);
    }

    sleep(3);

    char *testTx3 = QueryTXHash(testTx2);
    if (testTx3 == NULL)
    {
        printf("Failed to Test QueryTXHash\n");
        return 1;
    }
    else
    {
        printf("TxHash result %s\n", testTx3);
    }

    sleep(3);
    char *testTx4 = QueryWasmContract("nibi1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3slkhcux",
                                      "{\"owner_of\": {\"token_id\": \"unique-nft-15\", \"include_expired\": false}}");
    if (testTx4 == NULL)
    {
        printf("Failed to Test QueryWasmContract\n");
        return 1;
    }
    else
    {
        printf("Reponse %s\n", testTx4);
    }

    return 0;
}