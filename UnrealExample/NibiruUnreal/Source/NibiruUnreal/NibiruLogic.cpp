// Fill out your copyright notice in the Description page of Project Settings.

#include "NibiruLogic.h"

#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include <memory>
#include <string>
#include "unreal_nibi_sdk.h"

void UNibiruLogic::OnInitApp(bool &IsCreateOk, FString &error_return)
{
    //localnet
    IsCreateOk = false;
    int ret = NewNibiruClientDefault();
    if (ret != 0)
    {
        error_return = "Failed to create NibiruClient";
        printf("Failed to create NibiruClient\n");
        return;
    }
    IsCreateOk = true;
}

void UNibiruLogic::OnCreateWalletClicked(FString &address_key_return, bool &IsCreateOk, FString &error_return)
{
    IsCreateOk = false;
    error_return = "";
    address_key_return = "";
    char *keyName = strdup("TestKey");
    // Create new wallet
    // Generate Menomonic
    char *mnemonic = strdup("napkin rigid magnet grass plastic spawn replace hobby tray eternal pupil olive pledge nasty animal base bitter climb guess analyst fat neglect zoo earn");
    
    // Create key(private,public =>signner) from menemonic
    char *passPhase = strdup("pass");
    int createAccount = CreateAccount(keyName, mnemonic, passPhase);
    if (createAccount != 0)
    {
        error_return = "Failed to create account";
        printf("Failed to create account\n");
        return;
    }
    IsCreateOk = true;

    // Get account address
    char *address = GetAddressFromKeyName(keyName);
    printf("Account Address: %s\n", address);
    address_key_return = address;
}

 void UNibiruLogic::GetAccountBalance(FString address, FString &balance_return, bool &IsSuccess, FString &error_return){
    IsSuccess = false;
    auto convertedStr = StringCast<ANSICHAR>(*address);
    const char* queryAddress = convertedStr.Get();
    BaseAccount *account = QueryAccount((char*)queryAddress);
    balance_return="";
    //balance_return = "Balances :";
    if (account == NULL){
        error_return = "Failed to GetAccountBalance";
        return;
    }
    balance_return += std::to_string(account->Coins->Array[0].Amount).c_str();
    // for (int i=0; i < account->Coins->Length; i++){
    //     balance_return += account->Coins->Array[0].Denom;
    //     balance_return += " - ";
    //     balance_return += std::to_string(account->Coins->Array[0].Amount).c_str();
    //     balance_return += ";";
    // }
    IsSuccess = true;
 }

void UNibiruLogic::OnFaucetClicked(FString address_received, bool &IsSuccess, FString &error_return){
    error_return ="";
    IsSuccess = false;
    char *adminMnemonic = strdup("guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host");
    char *passPhase = strdup("pass");
    char *keyNameAdmin = strdup("AdminKey");
    int createAdminAccount = CreateAccount(keyNameAdmin, adminMnemonic, passPhase);
    if (createAdminAccount != 0)
    {
        error_return = "Failed to create admin account";
        printf("Failed to create account\n");
        return;
    }
    char *adminAddress = GetAddressFromKeyName(keyNameAdmin);
    printf("Admin Address: %s\n", adminAddress);

    auto convertedStr = StringCast<ANSICHAR>(*address_received);
    const char* toAddress = convertedStr.Get();
    char *demon = strdup("unibi");
    int tx = TransferToken(adminAddress, (char*)toAddress, demon, 250);
    delete[] toAddress;
    if (tx != 0)
    {
        error_return = "Failed to transfer";
        printf("Failed to transfer\n");
        return;
    }
    IsSuccess = true;
}


void UNibiruLogic::OnTransferClicked(FString from_address, FString to_address, FString demon, int amount, bool &IsSuccess, FString &error_return){
    IsSuccess = false;
    error_return = "";
    auto convertedStr = StringCast<ANSICHAR>(*from_address);
    const char* fromAddress = convertedStr.Get();
    auto convertedStr2 = StringCast<ANSICHAR>(*to_address);
    const char* toAddress = convertedStr2.Get();
    auto convertedStr3 = StringCast<ANSICHAR>(*demon);
    const char* demonStr = convertedStr3.Get();
    
    int tx = TransferToken((char*)fromAddress, (char*)toAddress, (char*)demonStr, amount);
    delete[] fromAddress;
    delete[] toAddress;
    delete[] demonStr;
    if (tx != 0)
    {
        error_return = "Failed to transfer";
        printf("Failed to transfer\n");
        return;
    }
    IsSuccess = true;
}