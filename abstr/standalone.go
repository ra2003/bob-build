/*
 * Copyright 2019 Arm Limited.
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

package abstr

import (
	"github.com/google/blueprint"
)

type visitableModuleContext interface {
	WalkDeps(func(blueprint.Module, blueprint.Module) bool)
	VisitDirectDepsIf(func(blueprint.Module) bool, func(blueprint.Module))
}

func TopDownAdaptor(f func(TopDownMutatorContext)) blueprint.TopDownMutator {
	return func(mctx blueprint.TopDownMutatorContext) {
		f(mctx)
	}
}

func BottomUpAdaptor(f func(BottomUpMutatorContext)) blueprint.BottomUpMutator {
	return func(mctx blueprint.BottomUpMutatorContext) {
		f(mctx)
	}
}

func Module(mctx BaseModuleContext) blueprint.Module {
	return mctx.(blueprint.BaseModuleContext).Module()
}

func WalkDeps(mctx VisitableModuleContext, f func(blueprint.Module, blueprint.Module) bool) {
	mctx.WalkDeps(f)
}

func VisitDirectDepsIf(mctx VisitableModuleContext, pred func(blueprint.Module) bool, f func(blueprint.Module)) {
	mctx.VisitDirectDepsIf(pred, f)
}

func CreateVariations(mctx BottomUpMutatorContext, variationNames ...string) []blueprint.Module {
	return mctx.(blueprint.BottomUpMutatorContext).CreateVariations(variationNames...)
}
