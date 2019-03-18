/*
 * Copyright 2018-2019 Arm Limited.
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

import (
	"path/filepath"

	"github.com/google/blueprint"

	"github.com/ARM-software/bob-build/utils"
)

// Support generation of static and shared libraries
// This file declares the common properties and functions needed by both.

// GenerateLibraryProps contain the properties that are specific to generating libraries
type GenerateLibraryProps struct {
	// List of headers that are created (if any)
	Headers []string
}

type generateLibrary struct {
	generateCommon
	Properties struct {
		GenerateLibraryProps
	}
}

// Modules implementing generateLibraryInterface support arbitrary commands
// that either produce a static library, shared library or binary.
type generateLibraryInterface interface {
	blueprint.Module
	dependentInterface

	libExtension() string
	getSources(ctx blueprint.ModuleContext) []string
}

//// Local functions

// Return a list of shared libraries generated by this module with full paths to the generated location
func getLibraryGeneratedPath(m generateLibraryInterface, g generatorBackend) string {
	return filepath.Join(m.outputDir(g), m.Name()+m.libExtension())
}

// Map sources to outputs. This is primarily to support
// transformSource, so here we return a single element associating all
// inputs with all outputs
func inouts(m generateLibraryInterface, ctx blueprint.ModuleContext) []inout {
	var io inout
	g := getBackend(ctx)
	io.srcIn = utils.PrefixDirs(m.getSources(ctx), g.sourcePrefix())
	io.genIn = getGeneratedFiles(ctx)
	io.out = m.outputs(g)
	return []inout{io}
}

//// Support Splittable

func (m *generateLibrary) supportedVariants() []string {
	return []string{m.generateCommon.Properties.Target}
}

func (m *generateLibrary) disable() {
	// This should never actually be called, as we will always support one target
	panic("disable() called on GenerateLibrary")
}

func (m *generateLibrary) setVariant(variant string) {
	// No need to actually track this, as a single target is always supported
}

func (m *generateLibrary) getSplittableProps() *SplittableProps {
	return &m.generateCommon.Properties.FlagArgsBuild.SplittableProps
}

func (m *generateLibrary) topLevelProperties() []interface{} {
	return append(m.generateCommon.topLevelProperties(), &m.Properties.GenerateLibraryProps)
}

// Support singleOutputModule interface
func (m *generateLibrary) outputName() string {
	return m.Name()
}
