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

    // "local" | "dev" | "test" | "main"
    char *networkType = "test";

    int switchNetwork = SwitchNetwork(networkType);

    char *keyName = "name";
    // Create new wallet
    // Generate Menomonic
    char *prases = GenerateRecoveryPhrase();

    // Create key(private,public =>signner) from menemonic
    char *passPrares = "pass";
    int createAccount = CreateAccount(keyName, prases, passPrares);
    if (createAccount != 0)
    {
        printf("Failed to create account\n");
        return 1;
    }

    u_int8_t *privkey = GetPrivKeyFromMnemonic(prases, keyName);
    if (privkey == NULL)
    {
        printf("Failed to get private key\n");
        return 1;
    }

    // int import = ImportAccountFromMnemoic(prases, keyName);

    // if (import != 0)
    // {
    //     printf("\nFailed to import account\n");
    //     return 1;
    // }

    // Get wallet address
    char *address = GetAddressFromKeyName(keyName);
    if (address == NULL)
    {
        printf("Failed to get Address key\n");
        return 1;
    }

    // int importP = ImportAccountFromPrivateKey(privkey, keyName);

    // if (importP != 0)
    // {
    //     printf("\nFailed to import account from privateKey\n");
    //     return 1;
    // }

    // int len;

    // KeyInfo **keyInfos = GetListAccount(&len);

    // if (keyInfos != NULL)
    // {
    //     // Use the keyInfos array
    //     for (int i = 0; i < len; ++i)
    //     {
    //         KeyInfo *keyInfo = keyInfos[i];
    //         // Do something with keyInfo, e.g., print it
    //         printf("Key Name: %s\n", keyInfo->Name);
    //         printf("Key Type: %d\n", keyInfo->Type);
    //         // printf("Key Address: %s\n", keyInfo->Address);
    //         // printf("Key PubKey: %s\n", keyInfo->PubKey);
    //     }
    // }

    // KeyInfo *keyInfo = GetAccountByKeyName(keyName);
    // if (keyInfo != NULL)
    // {
    //     printf("KeyInfo: %s\n", keyInfo->Name);
    // }

    KeyInfo *KeyInfo2 = GetAccountByAddress(address);
    if (KeyInfo2 != NULL)
    {
        printf("KeyInfo Address: %s\n", KeyInfo2->Address);
    }

    // int deleteAccount = DeleteAccount(keyName, passPrares);
    // if (deleteAccount != 0)
    // {
    //     printf("\nFailed to delete account\n");
    //     return 1;
    // }

    int testTx = TestTransferToken();
    if (testTx != 0)
    {
        printf("Failed to Test transfer\n");
        return 1;
    }

    // BaseAccount *baseAcc = QueryAccount(address);
    // if (baseAcc != NULL)
    // {
    //     printf("User coins count: %lu\n", baseAcc->Coins->Length);
    //     for (int i = 0; i < baseAcc->Coins->Length; i++)
    //     {
    //         printf("%d, %s coins have %llu\n", i + 1, baseAcc->Coins->Array[i].Denom, baseAcc->Coins->Array[i].Amount);
    //     }
    // }
    // else
    // {
    //     printf("\n Err: can't get base account");
    // }

    // return 0;
}