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
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func main() {
	// type t struct {
	// 	N int
	// }
	// var n = t{42}
	// // N at start
	// fmt.Println(n.N)
	// // pointer to struct - addressable
	// ps := reflect.ValueOf(&n)
	// // struct
	// s := ps.Elem()
	// if s.Kind() == reflect.Struct {
	// 	// exported field
	// 	f := s.FieldByName("N")
	// 	if f.IsValid() {
	// 		// A Value can be changed only if it is
	// 		// addressable and was not obtained by
	// 		// the use of unexported struct fields.
	// 		if f.CanSet() {
	// 			// change value of N
	// 			if f.Kind() == reflect.Int {
	// 				x := int64(7)
	// 				if !f.OverflowInt(x) {
	// 					f.SetInt(x)
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// // N at end
	// fmt.Println(n.N)
	type MyStruct struct {
		Name string
		Age  int64
	}

	myData := make(map[string]interface{})
	// myData["name"] = "Tony"
	myData["Age"] = 23

	result := MyStruct{
		Name: "Tong",
	}
	mapstructure.Decode(result, &myData)
	fmt.Print(result)
	// cmd.Execute()
}
