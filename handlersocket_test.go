/*
Copyright 2011 Brian Ketelsen

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

package handlersocket

import (
	"testing"
	"fmt"
)


func TestOpenIndex(t *testing.T) {

	// Create new instance
	hs := New()
	// Enable logging - which doesn't do much yet
	hs.Logging = true
	
	// Connect to database
	hs.Connect("192.168.1.120", 9998, 9999)
	defer hs.Close()

	hs.OpenIndex(1, "clarity_development", "users", "PRIMARY", "id", "login", "email")



}


func TestRead(t *testing.T) {
	hs := New()
	// Enable logging
	hs.Logging = true
	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()
	// Use UTF8
	hs.OpenIndex(1, "gotesting", "kvs", "PRIMARY", "id", "content")

	found, _ := hs.Find(1, "=", 1, 0, "brian")

	for i := range found {
		fmt.Println(found[i].Data)
	}

	fmt.Println(len(found), "rows returned")

}

func BenchmarkOpenIndex(b *testing.B) {
	b.StopTimer()
	hs := New()
	defer hs.Close()
	hs.Connect("192.168.1.120", 9998, 9999)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
	hs.OpenIndex(1, "gotesting", "kvs", "PRIMARY", "id", "content")

	}
}
func BenchmarkFind(b *testing.B) {

	b.StopTimer()
	hs := New()
	defer hs.Close()
	hs.Connect("192.168.1.120", 9998, 9999)
	hs.OpenIndex(1, "gotesting", "kvs", "PRIMARY", "id", "content")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		hs.Find(1, "=", 1, 0, "brian")
	}
}
