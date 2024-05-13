// Fill out your copyright notice in the Description page of Project Settings.

using UnrealBuildTool;
using System.IO;
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
		} else if (Target.Platform == UnrealTargetPlatform.Win64)
		{
			string unreal_nibi_sdk_LibPath = Path.Combine(ModuleDirectory, "../../../../", "unreal_nibi_sdk.dll");
			string cosmos_LibPath = Path.Combine(ModuleDirectory, "../../../../", "wasmvm.dll");
			string destinationDirectory = Target.ProjectFile.Directory.FullName + "/Binaries/Win64";
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
