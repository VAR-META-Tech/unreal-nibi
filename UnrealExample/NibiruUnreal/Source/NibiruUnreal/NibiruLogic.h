// Fill out your copyright notice in the Description page of Project Settings.

#pragma once
#define BOOST_NO_CXX98_FUNCTION_BASE
#include "CoreMinimal.h"
#include "Kismet/BlueprintFunctionLibrary.h"
#pragma GCC diagnostic ignored "-Wall"
#pragma GCC diagnostic ignored "-Wdeprecated-builtins"
#pragma GCC diagnostic ignored "-Wshadow"
#pragma GCC diagnostic ignored "-Wenum-constexpr-conversion"

#ifdef check
#undef check
#endif
#ifdef verify
#undef verify
#endif

#include "unreal_nibi_sdk.h"
#pragma GCC diagnostic pop

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
};
