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
	"net"
	"net/textproto"
	"os"
	"log"
	"io"
	"fmt"
	"strings"
	
)

type HandlerSocketError struct {
	Code	string
	Description	string
}

type HandlerSocketConnection struct {
	tcpConn        *net.TCPConn
	incomingChannel chan *HandlerSocketMessage
	outgoingChannel chan *HandlerSocketMessage
	logger	*log.Logger
	lastError	*HandlerSocketError
}


type HandlerSocketTarget struct {
	database string
	table	string
	indexname	string
	columns		[]string

}


/*
----------------------------------------------------------------------------
'open_index' request

The 'open_index' request has the following syntax.

    P <indexid> <dbname> <tablename> <indexname> <columns>

- <indexid> is a number in decimal.
- <dbname>, <tablename>, and <indexname> are strings. To open the primary
  key, use PRIMARY as <indexname>.
- <columns> is a comma-separated list of column names.

Once an 'open_index' request is issued, the HandlerSocket plugin opens the
specified index and keep it open until the client connection is closed. Each
open index is identified by <indexid>. If <indexid> is already open, the old
open index is closed. You can open the same combination of <dbname>
<tablename> <indexname> multple times, possibly with different <columns>.
For efficiency, keep <indexid> small as far as possible.

----------------------------------------------------------------------------
*/

func buildOpenIndexCommand(target HandlerSocketTarget ) (cmd string){

	cmd = ""
	cmd += "P"	
	cmd += "\t"
	cmd += "1"  //hack! ++ need something else like an auto incr or a hash with smarts
	cmd += "\t"
	cmd += target.database
	cmd += "\t"
	cmd += target.table
	cmd += "\t"
	cmd += target.indexname
	cmd += "\t"
	
	cmd += strings.Join(target.columns, ",")
	cmd += "\n"

	fmt.Println(cmd)
	return
}

func buildHandlerSocketError(response []byte, length int, action string) *HandlerSocketError{
	stringResponse := string(response[0:length])
	retVal := strings.Split(stringResponse,"\t",-1)
	hse := HandlerSocketError{Code:retVal[0], Description:action}
	return &hse
}

func (self *HandlerSocketConnection) OpenIndex(indexid int, target HandlerSocketTarget) {
		
		var command =[]byte(buildOpenIndexCommand(target))

		_,err := self.tcpConn.Write(command)
		if err != nil {
			self.lastError = &HandlerSocketError{Code:"-1",Description:"TCP Write Failed"}
			return
		}
		
		b := make([]byte, 256)
		m, err := self.tcpConn.Read(b)
		self.lastError = buildHandlerSocketError(b,m,"Open Index")


}

	
func (h HandlerSocketConnection) Close() (err os.Error) {
	if err := h.tcpConn.Close(); err != nil {
		return os.EINVAL
	}
	return nil
}

func NewHandlerSocketConnection(address string) *HandlerSocketConnection {
	
	localAddress, _ := net.ResolveTCPAddr("0.0.0.0:0")
	hsAddress, err := net.ResolveTCPAddr(address)

	if(err != nil) {
		return nil
	}

	tcpConn, err := net.DialTCP("tcp", localAddress, hsAddress)

	if(err != nil) {
		return nil
	}

	var newHsConn HandlerSocketConnection

	newHsConn.tcpConn = tcpConn
	newHsConn.incomingChannel = make(chan *HandlerSocketMessage, 100)
	newHsConn.outgoingChannel = make(chan *HandlerSocketMessage, 100)
	newHsConn.lastError = &HandlerSocketError{}

//	go newHsConn.Dispatch()

	return &newHsConn
}




