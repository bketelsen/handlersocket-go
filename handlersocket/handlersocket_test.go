// Copyright 2010  The "handlersocket-go" Authors
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

package handlersocket

import (
	"testing"
	"fmt"
//	"syscall"
)


func TestOpenIndex(t *testing.T) {
	
	if c := NewHandlerSocketConnection("127.0.0.1:9999"); c != nil{
		defer c.Close()
		target := HandlerSocketTarget{database:"hstest",table:"hstest_table1",indexname:"PRIMARY", columns:[]string{"k","v"}}
		c.OpenIndex(1,target)
		fmt.Println(c.lastError)
		if c.lastError.Code != "0" {
			t.Errorf("Last Error Code = %s, want %s.", c.lastError.Code, "0")
		}
	}

}

func TestWrite(t *testing.T) {
	
	if c := NewHandlerSocketConnection("127.0.0.1:9999"); c != nil{
		defer c.Close()
		target := HandlerSocketTarget{database:"hstest",table:"hstest_table1",indexname:"PRIMARY", columns:[]string{"k","v"}}
		c.OpenIndex(1,target)
		fmt.Println(c.lastError)
		if c.lastError.Code != "0" {
			t.Errorf("Last Error Code = %s, want %s.", c.lastError.Code, "0")
		}
	}
}
func TestRead(t *testing.T) {
	
	if c := NewHandlerSocketConnection("127.0.0.1:9999"); c != nil{
		defer c.Close()
		target := HandlerSocketTarget{database:"hstest",table:"hstest_table1",indexname:"PRIMARY", columns:[]string{"k","v"}}
		c.OpenIndex(1,target)
		fmt.Println(c.lastError)
		if c.lastError.Code != "0" {
			t.Errorf("Last Error Code = %s, want %s.", c.lastError.Code, "0")
		}
		c.Find("1","=","1","0","blue")
		
	}

}