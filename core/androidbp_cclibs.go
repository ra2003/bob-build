/*
 * Copyright 2020 Arm Limited.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package core

// Support for building libraries and binaries via soong's cc_library
// modules.

import (
	"fmt"

	"github.com/google/blueprint"
	"github.com/google/blueprint/proptools"

	"github.com/ARM-software/bob-build/internal/bpwriter"
	"github.com/ARM-software/bob-build/internal/utils"
)

// Convert between Bob module names, and the name we will give the generated
// cc_library module. This is required when a module supports being built on
// host and target; we cannot create two modules with the same name, so
// instead, we use the `shortName()` (which may include a `__host` or
// `__target` suffix) to disambiguate, and use the `stem` property to fix up
// the output filename.
func ccModuleName(mctx blueprint.BaseModuleContext, name string) string {
	var dep blueprint.Module

	mctx.VisitDirectDeps(func(m blueprint.Module) {
		if m.Name() == name {
			dep = m
		}
	})

	if dep == nil {
		panic(fmt.Errorf("%s has no dependency '%s'", mctx.ModuleName(), name))
	}

	if l, ok := getLibrary(dep); ok {
		return l.shortName()
	}

	// Most cases should match the getLibrary() check above, but generated libraries,
	// etc, do not, and they also do not require using shortName() (because of not
	// being target-specific), so just use the original build.bp name.
	return dep.Name()
}

func ccModuleNames(mctx blueprint.BaseModuleContext, nameLists ...[]string) []string {
	ccModules := []string{}
	for _, nameList := range nameLists {
		for _, name := range nameList {
			ccModules = append(ccModules, ccModuleName(mctx, name))
		}
	}
	return ccModules
}

func (l *library) getGeneratedSourceModules(mctx blueprint.BaseModuleContext) (srcs []string) {
	mctx.VisitDirectDepsIf(
		func(dep blueprint.Module) bool {
			return mctx.OtherModuleDependencyTag(dep) == generatedSourceTag
		},
		func(dep blueprint.Module) {
			switch dep.(type) {
			case *generateSource:
			case *transformSource:
			default:
				panic(fmt.Errorf("Dependency %s of %s is not a generated source",
					dep.Name(), l.Name()))
			}

			srcs = append(srcs, dep.Name())
		})
	return
}

func (l *library) getGeneratedHeaderModules(mctx blueprint.BaseModuleContext) (headers []string) {
	mctx.VisitDirectDepsIf(
		func(dep blueprint.Module) bool {
			return mctx.OtherModuleDependencyTag(dep) == generatedHeaderTag
		},
		func(dep blueprint.Module) {
			switch dep.(type) {
			case *generateSource:
			case *transformSource:
			default:
				panic(fmt.Errorf("Dependency %s of %s is not a generated source",
					dep.Name(), l.Name()))
			}

			headers = append(headers, dep.Name())
		})
	return
}

func addProvenanceProps(m bpwriter.Module, props AndroidProps) {
	if props.Owner != "" {
		m.AddString("owner", props.Owner)
		m.AddBool("vendor", true)
		m.AddBool("proprietary", true)
		m.AddBool("soc_specific", true)
	}
}

func addCcLibraryProps(m bpwriter.Module, l library, mctx blueprint.ModuleContext) {
	if len(l.Properties.Export_include_dirs) > 0 {
		panic(fmt.Errorf("Module %s exports non-local include dirs %v - this is not supported",
			mctx.ModuleName(), l.Properties.Export_include_dirs))
	}

	// Soong deals with exported include directories between library
	// modules, but it doesn't export cflags.
	_, _, exported_cflags := l.GetExportedVariables(mctx)

	cflags := utils.NewStringSlice(l.Properties.Cflags,
		l.Properties.Export_cflags, exported_cflags)

	sharedLibs := ccModuleNames(mctx, l.Properties.Shared_libs, l.Properties.Export_shared_libs)
	staticLibs := ccModuleNames(mctx, l.Properties.ResolvedStaticLibs)
	// Exported header libraries must be mentioned in both header_libs
	// *and* export_header_lib_headers - i.e., we can't export a header
	// library which isn't actually being used.
	headerLibs := ccModuleNames(mctx, l.Properties.Header_libs, l.Properties.Export_header_libs)

	reexportShared := []string{}
	reexportStatic := []string{}
	reexportHeaders := ccModuleNames(mctx, l.Properties.Export_header_libs)
	for _, lib := range ccModuleNames(mctx, l.Properties.Reexport_libs) {
		if utils.Contains(sharedLibs, lib) {
			reexportShared = append(reexportShared, lib)
		} else if utils.Contains(staticLibs, lib) {
			reexportStatic = append(reexportStatic, lib)
		} else if utils.Contains(headerLibs, lib) {
			reexportHeaders = append(reexportHeaders, lib)
		}
	}

	if l.shortName() != l.outputName() {
		m.AddString("stem", l.outputName())
	}
	m.AddStringList("srcs", utils.Filter(utils.IsCompilableSource, l.Properties.Srcs))
	m.AddStringList("generated_sources", l.getGeneratedSourceModules(mctx))
	m.AddStringList("generated_headers", l.getGeneratedHeaderModules(mctx))
	m.AddStringList("exclude_srcs", l.Properties.Exclude_srcs)
	m.AddStringList("cflags", cflags)
	m.AddStringList("include_dirs", l.Properties.Include_dirs)
	m.AddStringList("local_include_dirs", l.Properties.Local_include_dirs)
	m.AddStringList("shared_libs", ccModuleNames(mctx, l.Properties.Shared_libs, l.Properties.Export_shared_libs))
	m.AddStringList("static_libs", staticLibs)
	m.AddStringList("whole_static_libs", ccModuleNames(mctx, l.Properties.Whole_static_libs))
	m.AddStringList("header_libs", headerLibs)
	m.AddStringList("export_shared_lib_headers", reexportShared)
	m.AddStringList("export_static_lib_headers", reexportStatic)
	m.AddStringList("export_header_lib_headers", reexportHeaders)
	m.AddStringList("ldflags", l.Properties.Ldflags)
	if l.getInstallableProps().Relative_install_path != nil {
		m.AddString("relative_install_path", proptools.String(l.getInstallableProps().Relative_install_path))
	}

	addProvenanceProps(m, l.Properties.Build.AndroidProps)
}

func addStaticOrSharedLibraryProps(m bpwriter.Module, l library, mctx blueprint.ModuleContext) {
	// Soong's `export_include_dirs` field is relative to the module
	// dir. The Android.bp backend writes the file into the project
	// root, so we can use the Export_local_include_dirs property
	// unchanged.
	m.AddStringList("export_include_dirs", l.Properties.Export_local_include_dirs)
}

func addStripProp(m bpwriter.Module) {
	g := m.NewGroup("strip")
	g.AddBool("all", true)
}

func (g *androidBpGenerator) binaryActions(l *binary, mctx blueprint.ModuleContext) {
	if !enabledAndRequired(l) {
		return
	}

	var modType string
	switch l.Properties.TargetType {
	case tgtTypeHost:
		modType = "cc_binary_host"
	case tgtTypeTarget:
		modType = "cc_binary"
	}

	m, err := AndroidBpFile().NewModule(modType, l.shortName())
	if err != nil {
		panic(err.Error())
	}

	addCcLibraryProps(m, l.library, mctx)
	if l.strip() {
		addStripProp(m)
	}
}

func (g *androidBpGenerator) sharedActions(l *sharedLibrary, mctx blueprint.ModuleContext) {
	if !enabledAndRequired(l) {
		return
	}

	var modType string
	switch l.Properties.TargetType {
	case tgtTypeHost:
		modType = "cc_library_host_shared"
	case tgtTypeTarget:
		modType = "cc_library_shared"
	}

	m, err := AndroidBpFile().NewModule(modType, l.shortName())
	if err != nil {
		panic(err.Error())
	}

	addCcLibraryProps(m, l.library, mctx)
	addStaticOrSharedLibraryProps(m, l.library, mctx)
	if l.strip() {
		addStripProp(m)
	}
}

func (g *androidBpGenerator) staticActions(l *staticLibrary, mctx blueprint.ModuleContext) {
	if !enabledAndRequired(l) {
		return
	}

	var modType string
	switch l.Properties.TargetType {
	case tgtTypeHost:
		modType = "cc_library_host_static"
	case tgtTypeTarget:
		modType = "cc_library_static"
	}

	m, err := AndroidBpFile().NewModule(modType, l.shortName())
	if err != nil {
		panic(err.Error())
	}

	addCcLibraryProps(m, l.library, mctx)
	addStaticOrSharedLibraryProps(m, l.library, mctx)
}
