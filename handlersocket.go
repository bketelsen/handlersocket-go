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
	"net"
	"os"
	"log"
	"bufio"
	"sync"
	"fmt"
	"strings"
	"io"
	"strconv"
)

const (
	Version          = "0.0.3"
	DefaultReadPort  = 9998
	DefaultWritePort = 9999
	MaxPacketSize    = 1 << 24
)


/**
 * The main HandlerSocket struct
 * shamelessly modeled after Philio/GoMySQL
 * for consistency of usage
 */
type HandlerSocket struct {
	Logging bool
	auth    *HandlerSocketAuth
	conn    net.Conn
	wrConn  net.Conn
	//	In          <-chan HandlerSocketResponse
	in          chan HandlerSocketResponse
	out         chan HandlerSocketCommand
	wrIn        chan HandlerSocketResponse
	wrOut       chan HandlerSocketCommand
	connected   bool
	wrConnected bool
	mutex       *sync.Mutex
}


type HandlerSocketAuth struct {
	host      string
	dbname    string
	readPort  int
	writePort int
}

/**
 * Row definition
 */
type HandlerSocketRow struct {
	Data map[string]interface{}
}

type HandlerSocketCommand interface {
	writeTo(w io.Writer) (err os.Error)
}

type hsopencommand struct {
	command string
	params  []string
}

type hsfindcommand struct {
	command string
	params  []string
	limit   int
	offset  int
}

type hsmodifycommand struct {
	command  string
	criteria []string
	limit    int
	offset   int
	mop      string
	newvals  []string
}

type hsinsertcommand struct {
	command string
	params  []string
}

type HandlerSocketResponse struct {
	ReturnCode string
	Data       []string
}

type header map[string]string

var indexes map[int][]string

func (handlerSocket *HandlerSocket) OpenIndex(index int, dbName string, tableName string, indexName string, columns ...string) (err os.Error) {

	cols := strings.Join(columns, ",")
	strindex := strconv.Itoa(index)
	a := []string{strindex, dbName, tableName, indexName, cols}

	handlerSocket.mutex.Lock()
	handlerSocket.out <- &hsopencommand{command: "P", params: a}
	handlerSocket.wrOut <- &hsopencommand{command: "P", params: a}

	message2 := <-handlerSocket.wrIn
	message := <-handlerSocket.in

	handlerSocket.mutex.Unlock()

	indexes[index] = columns

	if message.ReturnCode != "0" {
		return os.NewError("Error Opening Index")
	}

	if message2.ReturnCode != "0" {
		return os.NewError("Error Opening Index")
	}

	return
}

/*

----------------------------------------------------------------------------
Updating/Deleting data

The 'find_modify' request has the following syntax.

    <indexid> <op> <vlen> <v1> ... <vn> <limit> <offset> <mop> <m1> ... <mk>

- <mop> is either 'U' (update) or 'D' (delete).
- <m1> ... <mk> specifies the column values to set. The length of <m1> ...
  <mk> must be smaller than or equal to the length of <columns> specified by
  the corresponding 'open_index' request. If <mop> is 'D', these parameters
  are ignored.
ind op	pc	key	lim off	mop	newpk	newval ...
1	=	1	red	1	0	U	red	brian
----------------------------------------------------------------------------

*/
func (handlerSocket *HandlerSocket) Modify(index int, oper string, limit int, offset int, modifyOper string, keys string, newvals string) (modifiedRows int, err os.Error) {

	//	query := strings.Join(keys, "\t")
	//	queryCount := strconv.Itoa(len(keys))

	//	a := []string{oper, queryCount, query}

	a := []string{oper, "1", keys}

	if modifyOper == "D" {

		handlerSocket.mutex.Lock()
		handlerSocket.wrOut <- &hsmodifycommand{command: strconv.Itoa(index), criteria: a, limit: limit, offset: offset, mop: modifyOper}
	}

	message := <-handlerSocket.wrIn
	handlerSocket.mutex.Unlock()

	if message.ReturnCode == "1" {
		return 0, os.NewError("Error Something")
	}

	return strconv.Atoi(strings.TrimSpace(message.Data[1]))

}


func (handlerSocket *HandlerSocket) Find(index int, oper string, limit int, offset int, vals ...string) (rows []HandlerSocketRow, err os.Error) {

	cols := strings.Join(vals, "\t")
	strindex := strconv.Itoa(index)
	colCount := strconv.Itoa(len(vals))
	a := []string{oper, colCount, cols}

	handlerSocket.mutex.Lock()
	handlerSocket.out <- &hsfindcommand{command: strindex, params: a, limit: limit, offset: offset}

	message := <-handlerSocket.in
	handlerSocket.mutex.Unlock()

	return parseResult(index, message), nil

}
/*
----------------------------------------------------------------------------
Inserting data

The 'insert' request has the following syntax.

    <indexid> '+' <vlen> <v1> ... <vn>

- <vlen> indicates the length of the trailing parameters <v1> ... <vn>. This
  must be smaller than or equal to the length of <columns> specified by the
  corresponding 'open_index' request.
- <v1> ... <vn> specify the column values to set. For columns not in
  <columns>, the default values for each column are set.

----------------------------------------------------------------------------
*/
func (handlerSocket *HandlerSocket) Insert(index int, vals ...string) (err os.Error) {

	cols := strings.Join(vals, "\t")
	strindex := strconv.Itoa(index)
	colCount := strconv.Itoa(len(vals))
	oper := "+"

	a := []string{oper, colCount, cols}

	handlerSocket.mutex.Lock()
	handlerSocket.wrOut <- &hsinsertcommand{command: strindex, params: a}
	message := <-handlerSocket.wrIn
	handlerSocket.mutex.Unlock()

	if message.ReturnCode == "1" {
		return os.NewError("Data Exists")
	}

	if message.ReturnCode != "0" {
		return os.NewError("Error Inserting Data")
	}
	return nil
}

func parseResult(index int, hs HandlerSocketResponse) (rows []HandlerSocketRow) {

	fieldCount, _ := strconv.Atoi(hs.Data[0])
	remainingFields := len(hs.Data) - 1
	if fieldCount > 0 {
		rs := remainingFields / fieldCount
		rows = make([]HandlerSocketRow, rs)

		offset := 1

		for r := 0; r < rs; r++ {
			d := make(map[string]interface{}, fieldCount)
			for f := 0; f < fieldCount; f++ {
				d[indexes[index][f]] = hs.Data[offset+f]
			}
			rows[r] = HandlerSocketRow{Data: d}
			offset += fieldCount
		}
	}
	return
}

/**
 * Close the connection to the server
 */
func (handlerSocket *HandlerSocket) Close() (err os.Error) {
	if handlerSocket.Logging {
		log.Print("Close called")
	}
	// If not connected return
	if !handlerSocket.connected {
		err = os.NewError("A connection to a MySQL server is required to use this function")
		return
	}

	if handlerSocket.Logging {
		log.Print("Sent quit command to server")
	}
	// Close connection
	handlerSocket.conn.Close()
	handlerSocket.connected = false
	if handlerSocket.Logging {
		log.Print("Closed connection to server")
	}
	return
}


/**
 * Reconnect (if connection droppped etc)
 */
func (handlerSocket *HandlerSocket) Reconnect() (err os.Error) {
	if handlerSocket.Logging {
		log.Print("Reconnect called")
	}

	// Close connection (force down)
	if handlerSocket.connected {
		handlerSocket.conn.Close()
		handlerSocket.connected = false
	}

	// Call connect
	err = handlerSocket.connect()
	return
}


/**
 * Connect to a server
 */
func (handlerSocket *HandlerSocket) Connect(params ...interface{}) (err os.Error) {
	if handlerSocket.Logging {
		log.Print("Connect called")
	}
	// If already connected return
	if handlerSocket.connected {
		err = os.NewError("Already connected to server")
		return
	}

	// Check min number of params
	if len(params) < 2 {
		err = os.NewError("A hostname and username are required to connect")
		return
	}
	// Parse params
	handlerSocket.parseParams(params)
	// Connect to server
	err = handlerSocket.connect()
	return
}

/**
 * Create a new instance of the package
 */
func New() (handlerSocket *HandlerSocket) {
	// Create and return a new instance of HandlerSocket
	handlerSocket = new(HandlerSocket)
	// Setup mutex
	handlerSocket.mutex = new(sync.Mutex)
	return
}

/**
 * Create connection to server using unix socket or tcp/ip then setup buffered reader/writer
 */
func (handlerSocket *HandlerSocket) connect() (err os.Error) {
	localAddress, _ := net.ResolveTCPAddr("0.0.0.0:0")
	targetAddress := fmt.Sprintf("%s:%d", handlerSocket.auth.host, handlerSocket.auth.readPort)
	wrTargetAddress := fmt.Sprintf("%s:%d", handlerSocket.auth.host, handlerSocket.auth.writePort)

	hsAddress, err := net.ResolveTCPAddr(targetAddress)
	hsWrAddress, err := net.ResolveTCPAddr(wrTargetAddress)

	handlerSocket.conn, err = net.DialTCP("tcp", localAddress, hsAddress)
	handlerSocket.wrConn, err = net.DialTCP("tcp", localAddress, hsWrAddress)

	if handlerSocket.Logging {
		log.Print("Connected using TCP/IP")
	}

	handlerSocket.in = make(chan HandlerSocketResponse)
	handlerSocket.out = make(chan HandlerSocketCommand)
	handlerSocket.wrIn = make(chan HandlerSocketResponse)
	handlerSocket.wrOut = make(chan HandlerSocketCommand)

	go handlerSocket.reader(handlerSocket.conn)
	go handlerSocket.writer(handlerSocket.conn)

	go handlerSocket.wrreader(handlerSocket.wrConn)
	go handlerSocket.wrwriter(handlerSocket.wrConn)

	indexes = make(map[int][]string, 10)

	handlerSocket.connected = true
	return
}


/**
 * Parse params given to Connect()
 */
func (handlerSocket *HandlerSocket) parseParams(p []interface{}) {
	handlerSocket.auth = new(HandlerSocketAuth)
	// Assign default values
	handlerSocket.auth.readPort = DefaultReadPort
	handlerSocket.auth.writePort = DefaultWritePort
	// Host / username are required
	handlerSocket.auth.host = p[0].(string)
	if len(p) > 1 {
		handlerSocket.auth.readPort = p[1].(int)
	}
	if len(p) > 3 {
		handlerSocket.auth.writePort = p[2].(int)
	}

	return
}


func (f *hsopencommand) writeTo(w io.Writer) os.Error {

	if _, err := fmt.Fprintf(w, "%s\t%s\n", f.command, strings.Join(f.params, "\t")); err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	return nil
}

func (f *hsfindcommand) writeTo(w io.Writer) os.Error {

	if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%d\n", f.command, strings.Join(f.params, "\t"), f.limit, f.offset); err != nil {
		return err
	}

	return nil
}

func (f *hsmodifycommand) writeTo(w io.Writer) os.Error {

	fmt.Printf("%s\t%s\t%d\t%d\t%s\n", f.command, strings.Join(f.criteria, "\t"), f.limit, f.offset, f.mop)

	if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n", f.command, strings.Join(f.criteria, "\t"), f.limit, f.offset, f.mop); err != nil {
		return err
	}

	return nil
}

func (f *hsinsertcommand) writeTo(w io.Writer) os.Error {

	if _, err := fmt.Fprintf(w, "%s\t%s\n", f.command, strings.Join(f.params, "\t")); err != nil {
		return err
	}

	return nil
}


func (c *HandlerSocket) reader(nc net.Conn) {
	br := bufio.NewReader(nc)
	var retString string
	for {

		b, err := br.ReadByte()
		if err != nil {
			// TODO(adg) handle error
			if err == os.EOF {
				break
			}
		}
		retString += string(b)
		if string(b) == "\n" {
			strs := strings.Split(retString, "\t", -1)
			hsr := HandlerSocketResponse{ReturnCode: strs[0], Data: strs[1:]}
			c.in <- hsr
			retString = ""
		}
	}
}

func (c *HandlerSocket) writer(nc net.Conn) {
	bw := bufio.NewWriter(nc)

	for f := range c.out {

		if err := f.writeTo(bw); err != nil {
			fmt.Println("ERROR:", err)
		}

		if err := bw.Flush(); err != nil {
			fmt.Println("ERROR:", err)
		}

	}
	nc.Close()
	c.connected = false
}

func (c *HandlerSocket) wrreader(nc net.Conn) {
	br := bufio.NewReader(nc)
	var retString string
	for {
		b, err := br.ReadByte()
		if err != nil {
			// TODO(adg) handle error
			if err == os.EOF {
				break
			}
		}
		retString += string(b)
		if string(b) == "\n" {
			strs := strings.Split(retString, "\t", -1)
			hsr := HandlerSocketResponse{ReturnCode: strs[0], Data: strs[1:]}
			c.wrIn <- hsr
			retString = ""
		}
	}
}

func (c *HandlerSocket) wrwriter(nc net.Conn) {
	bw := bufio.NewWriter(nc)

	for f := range c.wrOut {
		if err := f.writeTo(bw); err != nil {
			fmt.Println("ERROR:", err)
		}

		if err := bw.Flush(); err != nil {
			fmt.Println("ERROR:", err)
		}

	}
	nc.Close()
	c.connected = false
}
