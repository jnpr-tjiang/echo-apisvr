/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/jnpr-tjiang/echo-apisvr/cmd"
)

type test struct {
	msg string
}

func msg() string {
	return "hello"
}

func main() {
	// x := test{
	// 	msg: msg(),
	// }
	// fmt.Println(x)
	// var project models.Entity
	// project = &models.Project{}
	// t := reflect.TypeOf(project).Elem()
	// for i := 0; i < t.NumField(); i++ {
	// 	f := t.Field(i)
	// 	if f.Type.Name() != "BaseModel" && f.Type.Kind().String() == "slice" {
	// 		e := f.Type.Elem()
	// 		fmt.Println(e.Name())
	// 	}
	// }
	cmd.Execute()
}
