// Fill out your copyright notice in the Description page of Project Settings.

#include "NibiruLogic.h"

#include <stdio.h>
#include <stdlib.h>

#include "unreal_nibi_sdk.h"

void UNibiruLogic::OnInitApp()
{
    int ret = NewNibiruClientDefault();
    if (ret != 0)
    {
        printf("Failed to create NibiruClient\n");
    }
}