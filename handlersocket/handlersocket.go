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
	
)

const LF = 0x0a
const HT = 0x09



type HandlerSocketConnection struct {
	tcpConn        *net.TCPConn
	incomingChannel chan *HandlerSocketMessage
	outgoingChannel chan *HandlerSocketMessage
	logger	*log.Logger
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
func (h HandlerSocketConnection) OpenIndex(indexid int, dbname string, tablename string, indexname string, columns string) {
		
		var command =[]byte("P\t1\thstest\thstest_table1\tPRIMARY\tk,v\n")

		n,err := h.tcpConn.Write(command)
		fmt.Println(n, err)
		
		b := make([]byte, 1024)
		m, err := h.tcpConn.Read(b)
		fmt.Println(b, m, err)
		fmt.Println(string(b[0:m]))
		
		// we should get "0	1" if everything works
		
		

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

//	go newHsConn.Dispatch()

	return &newHsConn
}

type HandlerSocketMessage struct {
	raw string
	message string
}

func NewHandlerSocketMessage(line string) *HandlerSocketMessage {
	return &HandlerSocketMessage{raw: line}
}

func (conn *HandlerSocketConnection) Dispatch() {


}

func ReadLineIter(conn io.ReadWriteCloser) chan string {

	ch := make(chan string)
	textConn := textproto.NewConn(conn)

	go func() {
		for {
			line, err := textConn.ReadLine()

			if err != nil {
				break
			}
			ch <- line
		}
		close(ch)
	}()

	return ch
}


