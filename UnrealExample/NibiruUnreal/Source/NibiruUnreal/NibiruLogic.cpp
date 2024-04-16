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

void UNibiruLogic::OnCreateWalletClicked(FString &address_key_return, FString &admin_address_key_return, bool &IsCreateOk, FString &error_return)
{
    IsCreateOk = false;
    error_return = "";
    admin_address_key_return = "";
    address_key_return = "";

    char *keyNameAdmin = strdup("AdminKey");
    char *keyName = strdup("TestKey");
    // Create new wallet
    // Generate Menomonic
    char *prases = strdup("napkin rigid magnet grass plastic spawn replace hobby tray eternal pupil olive pledge nasty animal base bitter climb guess analyst fat neglect zoo earn");
    char *adminPhases = strdup("guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host");

    // Create key(private,public =>signner) from menemonic
    char *passPrares = strdup("pass");
    int createAdminAccount = CreateAccount(keyNameAdmin, adminPhases, passPrares);
    if (createAdminAccount != 0)
    {
        error_return = "Failed to create admin account";
        printf("Failed to create account\n");
        return;
    }

    int createAccount = CreateAccount(keyName, prases, passPrares);
    if (createAccount != 0)
    {
        error_return = "Failed to create account";
        printf("Failed to create account\n");
        return;
    }
    IsCreateOk = true;
    // Get wallet address
    // Get account address
    char *address = GetAddressFromKeyName(keyName);
    char *adminAddress = GetAddressFromKeyName(keyNameAdmin);

    printf("Admin Address: %s\n", adminAddress);
    printf("Account Address: %s\n", address);
    admin_address_key_return = adminAddress;
    address_key_return = address;
}
