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

/*
CREATE  TABLE `gotest`.`kvs` (
  `id` VARCHAR(255) NOT NULL ,
  `content` VARCHAR(255) NULL ,
  PRIMARY KEY (`id`) )
ENGINE = InnoDB;


*/

package handlersocket

import (
	"testing"
	"strconv"
)


func TestOpenIndex(t *testing.T) {

	// Create new instance
	hs := New()
	// Enable logging - which doesn't do much yet
	hs.Logging = true

	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()

	hs.OpenIndex(1, "gotest", "kvs", "PRIMARY", "id", "content")

}

func TestDelete(t *testing.T) {

	hs := New()
	// Enable logging
	hs.Logging = true
	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()
	// id is varchar(255), content is text
	hs.OpenIndex(3, "gotest", "kvs", "PRIMARY", "id", "content")

	var keys, newvals []string
	
        keys = make([]string,1)
	newvals = make([]string,0)

	for n:=1; n<10; n++ {
		keys[0] = "blue" + strconv.Itoa(n)
		_, err := hs.Modify(3, "=", 10, 0, "D", keys, newvals)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestWrite(t *testing.T) {

	hs := New()
	// Enable logging
	hs.Logging = true
	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()
	// id is varchar(255), content is text
	hs.OpenIndex(3, "gotest", "kvs", "PRIMARY", "id", "content")

	err := hs.Insert(3, "blue1", "a quick brown fox jumped over a lazy dog")
	if err != nil {
		// We receive an error if the PK already exists.  This might not be a real "fail". 
		// To test for sure, change the PK above before testing.

		//TODO: make a new PK each time.
		t.Error(err)
	}
	err = hs.Insert(3, "blue2", "a quick brown fox jumped over a lazy dog")
	if err != nil {
		// We receive an error if the PK already exists.  This might not be a real "fail". 
		// To test for sure, change the PK above before testing.

		//TODO: make a new PK each time.
		t.Error(err)
	}

}

func TestModify(t *testing.T) {

	hs := New()
	// Enable logging
	hs.Logging = true
	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()
	// id is varchar(255), content is text
	hs.OpenIndex(3, "gotest", "kvs", "PRIMARY", "id", "content")

	err := hs.Insert(3, "blue3", "a quick brown fox jumped over a lazy dog")
	if err != nil {
		// We receive an error if the PK already exists.  This might not be a real "fail". 
		// To test for sure, change the PK above before testing.

		//TODO: make a new PK each time.
		t.Error(err)
	}
	err = hs.Insert(3, "blue4", "a quick brown fox jumped over a lazy dog")
	if err != nil {
		// We receive an error if the PK already exists.  This might not be a real "fail". 
		// To test for sure, change the PK above before testing.

		//TODO: make a new PK each time.
		t.Error(err)
	}



	var keys, newvals []string
	
	keys = make([]string,1)
	newvals = make([]string,2)
	
	keys[0] = "blue3"
	newvals[0] = "blue7"
	newvals[1] = "some new thing"

	_, err = hs.Modify(3, "=", 1, 0, "U", keys, newvals)
	if err != nil {
		t.Error(err)
	}



	keys[0] = "blue4"
	newvals[0] = "blue5"
	newvals[1] = "My new value!"
	_, err = hs.Modify(3, "=", 1, 0, "U", keys, newvals)
	if err != nil {
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

	hs.OpenIndex(1, "gotest", "kvs", "PRIMARY", "id", "content")

	found, _ := hs.Find(1, "=", 1, 0, "blue7")


	if len(found) < 1 {
		t.Error("Expected one record for blue7")
	}


}

func BenchmarkOpenIndex(b *testing.B) {
	b.StopTimer()
	hs := New()
	defer hs.Close()
	hs.Connect("127.0.0.1", 9998, 9999)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hs.OpenIndex(1, "gotest", "kvs", "PRIMARY", "id", "content")

	}
}
func BenchmarkFind(b *testing.B) {

	b.StopTimer()
	hs := New()
	defer hs.Close()
	hs.Connect("127.0.0.1", 9998, 9999)
	hs.OpenIndex(1, "gotest", "kvs", "PRIMARY", "id", "content")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		hs.Find(1, "=", 1, 0, "brian")
	}
}
