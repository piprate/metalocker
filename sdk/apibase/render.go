// Copyright 2022 Piprate Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apibase

import (
	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/utils/jsonw"
)

var (
	jsonContentType = []string{"application/json; charset=utf-8"}
)

// JSON is a replacement for gin.Context.JSON. It uses the JSON wrapper defined in jsonw
// instead of the standard encoding.json, and also uses streaming.
func JSON(c *gin.Context, code int, obj any) {
	c.Status(code)

	header := c.Writer.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = jsonContentType
	}

	err := jsonw.Encode(obj, c.Writer)
	if err != nil {
		panic(err)
	}
}
