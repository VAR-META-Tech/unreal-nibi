#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "unreal_nibi_sdk.h"

int main()
{
    // Create a new NibiruClient instance using the exported Go function.
    int ret = NewNibiruClientDefault();
    if (ret != 0)
    {
        printf("Failed to create NibiruClient\n");
        return 1;
    }

    char *keyName = "name";
    // Create new wallet
    // Generate Menomonic
    char *prases = GenerateRecoveryPhrase();

    // Create key(private,public =>signner) from menemonic
    char *passPrares = "pass";
    int createAccount = CreateAccountV2(keyName, prases, passPrares);
    if (createAccount != 0)
    {
        printf("Failed to create account\n");
        return 1;
    }

    char *privkey = GetPrivKeyFromMnemonic(prases, keyName);
    if (privkey == NULL)
    {
        printf("Failed to get private key\n");
        return 1;
    }

    printf("Private Key: %s\n", privkey);

    // Get wallet address
    char *address = GetAddressFromMnemonic(prases, keyName);
    if (address == NULL)
    {
        printf("Failed to get Address key\n");
        return 1;
    }
    printf("Private Address: %s\n", address);

    int import = ImportAccountFromMnemoic(prases, keyName);

    if (import != 0)
    {
        printf("\nFailed to import account\n");
        return 1;
    }

    int len;

    KeyInfo **keyInfos = GetListAccount(&len);

    if (keyInfos != NULL)
    {
        // Use the keyInfos array
        for (int i = 0; i < len; ++i)
        {
            KeyInfo *keyInfo = keyInfos[i];
            // Do something with keyInfo, e.g., print it
            printf("Key Name: %s\n", keyInfo->Name);
            printf("Key Type: %d\n", keyInfo->Type);
            // printf("Key Address: %s\n", keyInfo->Address);
            // printf("Key PubKey: %s\n", keyInfo->PubKey);
        }
    }

    KeyInfo *keyInfo = GetAccountByKeyName(keyName);
    if (keyInfo != NULL)
    {
        printf("KeyInfo: %s\n", keyInfo->Name);
    }

    int deleteAccount = DeleteAccount(keyName, passPrares);
    if (deleteAccount != 0)
    {
        printf("\nFailed to delete account\n");
        return 1;
    }

    return 0;
}