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

		string destinationDirectory = Target.ProjectFile.Directory.FullName;
		PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../../") });
		PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../../libs/") });
		PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../../libs/calc") });
		PublicIncludePaths.AddRange(new string[] { Path.Combine(ModuleDirectory, "../../../../libs/cmdctx") });
		// PublicAdditionalLibraries.Add(Path.Combine(destinationDirectory, "unreal_nibi_sdk.so"));
		PublicIncludePaths.AddRange(new string[] { "/usr/local/include/" });
		//bUseRTTI = true;
		bEnableUndefinedIdentifierWarnings = false;
		CppStandard = CppStandardVersion.Cpp17;

	}
}
