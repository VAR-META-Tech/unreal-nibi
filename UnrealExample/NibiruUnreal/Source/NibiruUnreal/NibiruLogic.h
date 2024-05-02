// Fill out your copyright notice in the Description page of Project Settings.

#pragma once
#include "CoreMinimal.h"
#include "Kismet/BlueprintFunctionLibrary.h"

#include <string>
#include "NibiruLogic.generated.h"

/**
 *
 */
UCLASS()
class NIBIRUUNREAL_API UNibiruLogic : public UBlueprintFunctionLibrary
{
	GENERATED_BODY()
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "CopyCurrentWalletAdress"), Category = "NibiruLogic")
	static void CopyCurrentWalletAdress(FString StringToCopy);
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "OnInitApp"), Category = "NibiruLogic")
	static void OnInitApp(bool &IsCreateOk, FString &error_return);
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "OnCreateWalletClicked"), Category = "UIController")
	static void OnCreateWalletClicked(FString &address_key_return, bool &IsCreateOk, FString &error_return);
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "OnFaucetClicked"), Category = "UIController")
	static void OnFaucetClicked(FString address_received, bool &IsSuccess, FString &error_return);
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "OnTransferClicked"), Category = "UIController")
	static void OnTransferClicked(FString from_address, FString to_address, FString demon, int amount, bool &IsSuccess, FString &error_return);
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "GetAccountBalance"), Category = "UIController")
	static void GetAccountBalance(FString address, FString &balance_return, bool &IsSuccess, FString &error_return);
};
