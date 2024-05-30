// Fill out your copyright notice in the Description page of Project Settings.

#include "NibiruLogic.h"
#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include <memory>
#include <string>
#ifdef _MSC_VER
#include <complex.h>
typedef std::complex<double>  _Dcomplex;
typedef std::complex<float>   _Fcomplex;
#endif
#include "../../../../unreal_nibi_sdk.h"


extern "C" {
    typedef int (*NewNibiruClientDefaultPtr)(); // Create a function pointer type
    typedef int (*CreateAccountPtr)(char*, char*, char*); // Create a function pointer type
    typedef char* (*GetAddressFromKeyNamePtr)(char*); // Create a function pointer type
    typedef BaseAccount* (*QueryAccountPtr)(char*); // Create a function pointer type
    typedef char* (*TransferTokenPtr)(char*, char*, char*, int); // Create a function pointer type
    typedef int (*SetLogFilePtr)(char*); // Create a function pointer type
}

void UNibiruLogic::CopyCurrentWalletAdress(FString StringToCopy)
{
  //  FPlatformMisc::ClipboardCopy(* StringToCopy);
}

void UNibiruLogic::OnInitApp(bool &IsCreateOk, FString &error_return)
{
    //testnet
#ifdef _MSC_VER
    void* DllHandle = FPlatformProcess::GetDllHandle(TEXT("unreal_nibi_sdk.dll"));
    NewNibiruClientDefaultPtr NewNibiruClientDefault = (NewNibiruClientDefaultPtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("NewNibiruClientDefault")));
    if (NewNibiruClientDefault == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find NewNibiruClientDefault function in DLL";
        printf("Failed to find NewNibiruClientDefault function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
    SetLogFilePtr SetLogFile = (SetLogFilePtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("SetLogFile")));
    if (SetLogFile == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find SetLogFile function in DLL";
        printf("Failed to find SetLogFile function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
    SetLogFile((char*)"D://unreal_nibi_sdk.log");
#endif
    IsCreateOk = false;
    int ret = NewNibiruClientDefault();
    if (ret != 0)
    {
        error_return = "Failed to create NibiruClient";
        printf("Failed to create NibiruClient\n");
        return;
    }
    IsCreateOk = true;
    error_return = "Successfully created NibiruClient.";
    //FPlatformProcess::FreeDllHandle(DllHandle);
}

void UNibiruLogic::OnCreateWalletClicked(FString &address_key_return, bool &IsCreateOk, FString &error_return)
{
#ifdef _MSC_VER
    void* DllHandle = FPlatformProcess::GetDllHandle(TEXT("unreal_nibi_sdk.dll"));
    CreateAccountPtr CreateAccount = (CreateAccountPtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("CreateAccount")));
    if (CreateAccount == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find CreateAccount function in DLL";
        printf("Failed to find CreateAccount function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
    GetAddressFromKeyNamePtr GetAddressFromKeyName = (GetAddressFromKeyNamePtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("GetAddressFromKeyName")));
    if (GetAddressFromKeyName == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find GetAddressFromKeyName function in DLL";
        printf("Failed to find GetAddressFromKeyName function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
#endif
    IsCreateOk = false;
    error_return = "";
    address_key_return = "";
    char *keyName = strdup("TestKey");
    // Create new wallet
    // Generate Menomonic
    char *mnemonic = strdup("toe cream coach quiz cactus nest spike gauge opinion legal father stadium lizard match wood immune odor depart sauce timber crash pig thought seat");
    
    // Create key(private,public =>signner) from menemonic
    char *passPhase = strdup("");
    int createAccount = CreateAccount(keyName, mnemonic, passPhase);
    if (createAccount != 0)
    {
        error_return = "Failed to create account";
        printf("Failed to create account\n");
        return;
    }
    // Get account address
    char *address = GetAddressFromKeyName(keyName);
    printf("Account Address: %s\n", address);
    address_key_return = address;

    IsCreateOk = true;
    error_return =  "Wallet created successfully";
}

 void UNibiruLogic::GetAccountBalance(FString address, FString &balance_return, bool &IsSuccess, FString &error_return){
#ifdef _MSC_VER
     void* DllHandle = FPlatformProcess::GetDllHandle(TEXT("unreal_nibi_sdk.dll"));
     QueryAccountPtr QueryAccount = (QueryAccountPtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("QueryAccount")));
     if (QueryAccount == nullptr) {
         // Handle error, function not found
         error_return = "Failed to find QueryAccount function in DLL";
         printf("Failed to find QueryAccount function in DLL\n");
         FPlatformProcess::FreeDllHandle(DllHandle);
         return;
     }
#endif
    IsSuccess = false;
    auto convertedStr = StringCast<ANSICHAR>(*address);
    const char* queryAddress = convertedStr.Get();
    BaseAccount *account = QueryAccount((char*)queryAddress);
    error_return="";
    balance_return="";
    //balance_return = "Balances :";
    if (account == NULL){
        error_return = "Failed to GetAccountBalance";
        return;
    }
    balance_return = std::to_string(account->Coins->Array[0].Amount).c_str();
    // for (int i=0; i < account->Coins->Length; i++){
    //     balance_return += account->Coins->Array[0].Denom;
    //     balance_return += " - ";
    //     balance_return += std::to_string(account->Coins->Array[0].Amount).c_str();
    //     balance_return += ";";
    // }
    IsSuccess = true;
    error_return="Successfully retrieved Account Balance";
 }

void UNibiruLogic::OnFaucetClicked(FString address_received, bool &IsSuccess, FString &error_return){
#ifdef _MSC_VER
    void* DllHandle = FPlatformProcess::GetDllHandle(TEXT("unreal_nibi_sdk.dll"));
    CreateAccountPtr CreateAccount = (CreateAccountPtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("CreateAccount")));
    if (CreateAccount == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find CreateAccount function in DLL";
        printf("Failed to find CreateAccount function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
    TransferTokenPtr TransferToken = (TransferTokenPtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("TransferToken")));
    if (TransferToken == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find TransferToken function in DLL";
        printf("Failed to find TransferToken function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
    GetAddressFromKeyNamePtr GetAddressFromKeyName = (GetAddressFromKeyNamePtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("GetAddressFromKeyName")));
    if (GetAddressFromKeyName == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find GetAddressFromKeyName function in DLL";
        printf("Failed to find GetAddressFromKeyName function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
#endif
    error_return ="";
    IsSuccess = false;
    char *adminMnemonic = strdup("guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host");
    char *passPhase = strdup("");
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
    char* tx = TransferToken(adminAddress, (char*)toAddress, demon, 11000000);
    if (tx == NULL)
    {
        error_return = "Failed to transfer";
        printf("Failed to transfer\n");
        return;
    }
    IsSuccess = true;
    error_return="Faucet Successfully!";
}


void UNibiruLogic::OnTransferClicked(FString from_address, FString to_address, FString demon, int amount, bool &IsSuccess, FString &error_return){
#ifdef _MSC_VER
    void* DllHandle = FPlatformProcess::GetDllHandle(TEXT("unreal_nibi_sdk.dll"));
    TransferTokenPtr TransferToken = (TransferTokenPtr)FPlatformProcess::GetDllExport(DllHandle, (TEXT("TransferToken")));
    if (TransferToken == nullptr) {
        // Handle error, function not found
        error_return = "Failed to find TransferToken function in DLL";
        printf("Failed to find TransferToken function in DLL\n");
        FPlatformProcess::FreeDllHandle(DllHandle);
        return;
    }
#endif
    IsSuccess = false;
    error_return = "";
    auto convertedStr = StringCast<ANSICHAR>(*from_address);
    const char* fromAddress_ = convertedStr.Get();
    convertedStr = StringCast<ANSICHAR>(*to_address);
    const char* toAddress_ = convertedStr.Get();
    convertedStr = StringCast<ANSICHAR>(*demon);
    const char* demonStr_ = convertedStr.Get();
    
    char* tx = TransferToken((char*)fromAddress_, (char*)toAddress_, (char*)demonStr_, amount);

   
    if (tx == NULL)
    {
        error_return = "Failed to transfer";
        printf("Failed to transfer\n");
        return;
    }
   
    IsSuccess = true;
    error_return = "Transfer Successfully ";
}