// Copyright 2014 The fader authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fader

import (
	"time"
)

// This interface has to be implemented by any item that should be handled by
// the fader. The Key() function should return a string that identified the item.
// The provided Time() will help the fader to schedule the expiry of the item.
type Item interface {
	Key() string
	Time() time.Time
}
