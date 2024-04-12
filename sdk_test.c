#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "unreal_nibi_sdk.h"
#include <unistd.h>

int main()
{
    // Create a new NibiruClient instance using the exported Go function.
    int ret = NewNibiruClientDefault();
    if (ret != 0)
    {
        printf("Failed to create NibiruClient\n");
        return 1;
    }

    char *keyNameAdmin = "AdminKey";
    char *keyName = "TestKey";
    // Create new wallet
    // Generate Menomonic
    char *prases = "napkin rigid magnet grass plastic spawn replace hobby tray eternal pupil olive pledge nasty animal base bitter climb guess analyst fat neglect zoo earn";
    char *adminPhases = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host";

    // Create key(private,public =>signner) from menemonic
    char *passPrares = "pass";
    int createAdminAccount = CreateAccount(keyNameAdmin, adminPhases, passPrares);
    if (createAdminAccount != 0)
    {
        printf("Failed to create account\n");
        return 1;
    }
    int createAccount = CreateAccount(keyName, prases, passPrares);
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

    // int testTx = TransferToken(adminAddress, address, "unibi", 250);
    // if (testTx != 0)
    // {
    //     printf("Failed to Test transfer\n");
    //     return 1;
    // }

    int testTx2 = ExecuteWasmContract(adminAddress, "nibi1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqugq26k",
    "{\"mint\": {\"token_id\": \"unique-nft-3\", \"owner\": \"nibi1zy7amen6h5e4whcta4ac656l0whsalzmnqrkc5\", \"token_uri\": \"https://metadata.com/nft1.json\"}}", 
    "unibi", 1);
    if (testTx2 != 0)
    {
        printf("Failed to Test transfer\n");
        return 1;
    }

    // int testTx3 = ExecuteWasmContract(adminAddress, "nibi1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqugq26k",
    // "{\"owner_of\": {\"token_id\": \"unique-nft-1\", \"include_expired\": false}}", 
    // "unibi", 0);
    // if (testTx3 != 0)
    // {
    //     printf("Failed to Test transfer\n");
    //     return 1;
    // }
    return 0;
}