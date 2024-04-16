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
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "OnInitApp"), Category = "NibiruLogic")
	static void OnInitApp();
	UFUNCTION(BlueprintCallable, meta = (DisplayName = "OnCreateWalletClicked"), Category = "UIController")
	static void OnCreateWalletClicked(FString &address_key_return, FString &admin_address_key_return, bool &IsCreateOk, FString &error_return);
};
