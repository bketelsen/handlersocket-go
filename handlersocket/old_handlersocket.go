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
	//	"net/textproto"
	"os"
	"log"
	//	"io"
	"bufio"
//	"bytes"
	"fmt"
	"strings"
)

type HandlerSocketError struct {
	Code        string
	Description string
}

type HandlerSocketConnection struct {
	tcpConn         *net.TCPConn
	reader 			*bufio.Reader
	writer 			*bufio.Writer 
	logger          *log.Logger
	lastError       *HandlerSocketError
	target			*HandlerSocketTarget
}


type HandlerSocketTarget struct {
	host	string
	port	int
	index	int
	database  string
	table     string
	indexname string
	columns   []string
}

var indexes map[int]HandlerSocketTarget

func init(){
	indexes = make(map[int]HandlerSocketTarget,10) //map of indexes
}
/*
---------------------------------------------------------------------------
Getting data

The 'find' request has the following syntax.

    <indexid> <op> <vlen> <v1> ... <vn> <limit> <offset>

- <indexid> is a number. This number must be an <indexid> specified by a
  'open_index' request executed previously on the same connection.
- <op> specifies the comparison operation to use. The current version of
  HandlerSocket supports '=', '>', '>=', '<', and '<='.
- <vlen> indicates the length of the trailing parameters <v1> ... <vn>. This
  must be smaller than or equal to the number of index columns specified by
  specified by the corresponding 'open_index' request.
- <v1> ... <vn> specify the index column values to fetch.
- <limit> and <offset> are numbers. These parameters can be omitted. When
  omitted, it works as if 1 and 0 are specified.

----------------------------------------------------------------------------
*/

func buildFindCommand(indexid string, operator string,  limit string, offset string, columns... string) (cmd string){

	cmd = fmt.Sprintf("%s\t%s\t%d\t%s\n", indexid, operator, len(columns),strings.Join(columns,"\t"))

	return
	
}

func (self *HandlerSocketConnection) Find(indexid string, operator string,  limit string, offset string, columns... string) () {
	// assumes the existence of an opened index
	
	fmt.Println("find")
	var command = []byte(buildFindCommand(indexid, operator,  limit, offset, columns...))
	_, err := self.tcpConn.Write(command)
	if err != nil {
		fmt.Println("err!")
		
		self.lastError = &HandlerSocketError{Code: "-2", Description: "TCP Write Failed"}
		return
	}
	

	space := []byte{'\t'}
	fmt.Println("-",string(space),"-")
	resultCode, err := self.reader.ReadBytes(space[0])
	if err != nil {
		fmt.Println("err!")
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}
	fmt.Println("returned:",string(resultCode))

//	self.lastError = buildHandlerSocketError(b, m, "Find")

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

func buildOpenIndexCommand(target HandlerSocketTarget) (cmd string) {

	cmd = fmt.Sprintf("P\t%d\t%s\t%s\t%s\t%s\n", target.index, target.database,target.table,target.indexname, strings.Join(target.columns,","))
	return
}

func buildHandlerSocketError(response []byte, length int, action string) *HandlerSocketError {
	stringResponse := string(response[0:length])
	retVal := strings.Split(stringResponse, "\t", -1)
	hse := HandlerSocketError{Code: retVal[0], Description: action}
	return &hse
}

func (self *HandlerSocketConnection) OpenIndex(target HandlerSocketTarget) {

	var command = []byte(buildOpenIndexCommand(target))
	fmt.Println("open")
	_, err := self.tcpConn.Write(command)
	if err != nil {
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP Write Failed"}
		return
	}

	newline := []byte{'\n'}
	tab := []byte{'\t'}
		fmt.Println("red",string(tab),"green",string(newline))
	success, err := self.reader.ReadByte()
	if err != nil {
		fmt.Println("err1!")
		
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}
	fmt.Println("success",string(success))
	
	delim, err0 := self.reader.ReadByte()
	if err0 != nil {
		fmt.Println("err0	!")
		
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}
	fmt.Println("delim",string(delim))
	
	
	code, err1 := self.reader.ReadByte()
	if err1 != nil {
		fmt.Println("err2	!")
		
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}
	fmt.Println("code",string(code))
	
//	self.lastError = buildHandlerSocketError(b, m, "Open Index")
	indexes[target.index] = target

}


func (h HandlerSocketConnection) Close() (err os.Error) {
	if err := h.tcpConn.Close(); err != nil {
		return os.EINVAL
	}
	return nil
}

func NewHandlerSocketConnection(target HandlerSocketTarget) *HandlerSocketConnection {

	localAddress, _ := net.ResolveTCPAddr("0.0.0.0:0")
	targetAddress := fmt.Sprintf("%s:%d",target.host, target.port)
	hsAddress, err := net.ResolveTCPAddr(targetAddress)

	if err != nil {
		return nil
	}

	tcpConn, err := net.DialTCP("tcp", localAddress, hsAddress)

	if err != nil {
		return nil
	}

	var newHsConn HandlerSocketConnection

	newHsConn.tcpConn = tcpConn
	newHsConn.lastError = &HandlerSocketError{}
	newHsConn.target = &target
	newHsConn.reader = bufio.NewReader(newHsConn.tcpConn)
	newHsConn.writer = bufio.NewWriter(newHsConn.tcpConn)
	

	//	go newHsConn.Dispatch()

	return &newHsConn
}

type HandlerSocketMessage struct {
	raw     string
	message string
}
