// Fill out your copyright notice in the Description page of Project Settings.

#include "NibiruLogic.h"

#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include "unreal_nibi_sdk.h"

void UNibiruLogic::OnInitApp()
{
    int ret = NewNibiruClientDefault();
    if (ret != 0)
    {
        printf("Failed to create NibiruClient\n");
    }
}

void UNibiruLogic::OnCreateWalletClicked(FString &menomonic_key_return, FString &privkey_key_return, FString &adress_key_return, bool &IsCreateOk)
{
    IsCreateOk = false;
    menomonic_key_return = "";
    privkey_key_return = "";
    adress_key_return = "";

    char *keyName = strdup("name");
    // Create new wallet
    // Generate Menomonic
    char *menomonic = GenerateRecoveryPhrase();
    printf("Prases: %s", menomonic);

    // Create key(private,public =>signner) from menemonic
    int createAccount = CreateAccount(keyName, menomonic);
    if (createAccount != 0)
    {
        printf("Failed to create account\n");
        return;
    }

    // Get wallet address
    int address = GetAddressFromMnemonic(menomonic, keyName);
    if (address != 0)
    {
        printf("Failed to get private key\n");
        return;
    }

    int privkey = GetPrivKeyFromMnemonic(menomonic, keyName);
    if (privkey != 0)
    {
        printf("Failed to get private key\n");
        return;
    }
}
