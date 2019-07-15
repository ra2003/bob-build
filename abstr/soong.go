// +build soong

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

/*
 * This file is included when Bob is being run as a Soong plugin.
 *
 * The build tag on the first line ensures that it is not included in the build
 * by accident, and that it is not included in `go test` or similar checks,
 * which would fail, because Soong is not available in that environment.
 */

package abstr

import (
	"github.com/google/blueprint"

	"android/soong/android"
)

func TopDownAdaptor(f func(TopDownMutatorContext)) android.AndroidTopDownMutator {
	return func(mctx android.TopDownMutatorContext) {
		f(mctx)
	}
}

func BottomUpAdaptor(f func(BottomUpMutatorContext)) android.AndroidBottomUpMutator {
	return func(mctx android.BottomUpMutatorContext) {
		f(mctx)
	}
}

func Module(mctx TopDownMutatorContext) blueprint.Module {
	return mctx.(android.TopDownMutatorContext).Module()
}