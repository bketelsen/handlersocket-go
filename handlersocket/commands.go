package handlersocket

type CommandStrings string

const (
	COM_OPENINDEX = "P"
)

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

/*

func buildOpenIndexCommand(target HandlerSocketTarget) (cmd string) {

	cmd = fmt.Sprintf("P\t%d\t%s\t%s\t%s\t%s\n", target.index, target.database, target.table, target.indexname, strings.Join(target.columns, ","))
	return
}


func (self *HandlerSocketConnection) OpenIndex(target HandlerSocketTarget) {

	var command = []byte(buildOpenIndexCommand(target))
	_, err := self.tcpConn.Write(command)
	if err != nil {
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP Write Failed"}
		return
	}

	_, err = self.reader.ReadByte()
	if err != nil {
		fmt.Println("err1!")
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}

	_, err0 := self.reader.ReadByte()
	if err0 != nil {
		fmt.Println("err0	!")
		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}

	_, err1 := self.reader.ReadByte()
	if err1 != nil {
		fmt.Println("err2	!")

		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}

	_, err2 := self.reader.ReadByte()
	if err2 != nil {
		fmt.Println("err2	!")

		self.lastError = &HandlerSocketError{Code: "-1", Description: "TCP read byte conversion failed"}
		return
	}

	self.lastError = &HandlerSocketError{Code: "0", Description: "SUCCESS"}

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
	targetAddress := fmt.Sprintf("%s:%d", target.host, target.port)
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

