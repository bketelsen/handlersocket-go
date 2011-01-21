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

	fmt.Println("Testing Open")
	// Create new instance
	hs := New()
	// Enable logging - which doesn't do much yet
	hs.Logging = true
	
	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()

	hs.OpenIndex(1, "clarity_development", "users", "PRIMARY", "id", "login", "email")



}

func TestWrite(t *testing.T) {
	
	fmt.Println("Testing Write")
	hs := New()
	// Enable logging
	hs.Logging = true
	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()
	// id is varchar(255), content is text
	hs.OpenIndex(3, "gotesting", "kvs", "PRIMARY", "id", "content")
	
	err := hs.Insert(3,"blue1","a quick brown fox jumped over a lazy dog")
	if err != nil {
		// We receive an error if the PK already exists.  This might not be a real "fail". 
		// To test for sure, change the PK above before testing.
		
		//TODO: make a new PK each time.
		t.Error(err)
		
	}
	
}


func TestRead(t *testing.T) {
	
	hs := New()
	// Enable logging
	hs.Logging = true
	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()

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
	hs.Connect("127.0.0.1", 9998, 9999)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
	hs.OpenIndex(1, "gotesting", "kvs", "PRIMARY", "id", "content")

	}
}
func BenchmarkFind(b *testing.B) {

	b.StopTimer()
	hs := New()
	defer hs.Close()
	hs.Connect("127.0.0.1", 9998, 9999)
	hs.OpenIndex(1, "gotesting", "kvs", "PRIMARY", "id", "content")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		hs.Find(1, "=", 1, 0, "brian")
	}
}
