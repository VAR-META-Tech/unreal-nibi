// Fill out your copyright notice in the Description page of Project Settings.

using UnrealBuildTool;
using System.IO;
using System;
public class NibiruUnreal : ModuleRules
{
	public NibiruUnreal(ReadOnlyTargetRules Target) : base(Target)
	{
		PCHUsage = PCHUsageMode.UseExplicitOrSharedPCHs;
		bEnableExceptions = true;
		PublicDependencyModuleNames.AddRange(new string[] { "Core", "CoreUObject", "Engine", "InputCore" });
		PrivateDependencyModuleNames.AddRange(new string[] { "Slate", "SlateCore" });
		PrivateDependencyModuleNames.AddRange(new string[] { });

		if (Target.Platform == UnrealTargetPlatform.Mac)
		{
			string unreal_nibi_sdk_LibPath = Path.Combine(ModuleDirectory, "../../../../", "unreal_nibi_sdk.dylib");
			string destinationDirectory = Target.ProjectFile.Directory.FullName;
			File.Copy(unreal_nibi_sdk_LibPath, Path.Combine(destinationDirectory, "unreal_nibi_sdk.dylib"), true);
			PublicAdditionalLibraries.Add(Path.Combine(destinationDirectory, "unreal_nibi_sdk.dylib"));
			PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../") });
			PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../../") });
		}
		else if (Target.Platform == UnrealTargetPlatform.Linux)
		{
			string unreal_nibi_sdk_LibPath = Path.Combine(ModuleDirectory, "../../../../", "unreal_nibi_sdk.so");
			string destinationDirectory = Path.Combine(Target.ProjectFile.Directory.FullName, "Binaries/Linux/");
			File.Copy(unreal_nibi_sdk_LibPath, Path.Combine(destinationDirectory, "unreal_nibi_sdk.so"), true);
			PublicAdditionalLibraries.Add(Path.Combine(destinationDirectory, "unreal_nibi_sdk.so"));
			PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../") });
			PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../../") });
		} 
		else if (Target.Platform == UnrealTargetPlatform.Win64)
		{
			string unreal_nibi_sdk_LibPath = Path.Combine(ModuleDirectory, "../../../../", "unreal_nibi_sdk.dll");
			string cosmos_LibPath = Path.Combine(ModuleDirectory, "../../../../", "wasmvm.dll");
			string destinationDirectory = Target.ProjectFile.Directory.FullName + "/Binaries/Win64";
            if (Directory.Exists(destinationDirectory))
            {
                Console.WriteLine("That path exists already.");
                return;
            }

            // Try to create the directory.
            DirectoryInfo di = Directory.CreateDirectory(destinationDirectory);
            File.Copy(unreal_nibi_sdk_LibPath, Path.Combine(destinationDirectory, "unreal_nibi_sdk.dll"), true);
			File.Copy(cosmos_LibPath, Path.Combine(destinationDirectory, "wasmvm.dll"), true);
            //PublicDelayLoadDLLs.Add(Path.Combine(destinationDirectory, "unreal_nibi_sdk.dll"));
            //PublicDelayLoadDLLs.Add(Path.Combine(destinationDirectory, "wasmvm.dll"));
            PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../") });
			PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../../") });
		}
		CppStandard = CppStandardVersion.Cpp17;
	}
}
