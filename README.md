handlersocket-go
================

Go library for connecting to HandlerSocket Mysql plugin.  See github.com/ahiguti/HandlerSocket-Plugin-for-MySQL/


## Installation

	$ go get github.com/bketelsen/handlersocket-go


## Read Example  - Best examples are in the TEST file.

	hs := New()

	// Connect to database
	hs.Connect("127.0.0.1", 9998, 9999)
	defer hs.Close()
	hs.OpenIndex(1, "gotesting", "kvs", "PRIMARY", "id", "content")

	found, _ := hs.Find(1, "=", 1, 0, "brian")

	for i := range found {
			fmt.Println(found[i].Data) 
		}

	fmt.Println(len(found), "rows returned")


## Write Example

	hs := New()
	hs.Connect("127.0.0.1", 9998, 9999) // host, read port, write port
	defer hs.Close()

	// id is varchar(255), content is text
	hs.OpenIndex(3, "gotesting", "kvs", "PRIMARY", "id", "content")

	err := hs.Insert(3,"mykey1","a quick brown fox jumped over a lazy dog")


## Modify Example

	var keys, newvals []string
	keys = make([]string,1)
	newvals = make([]string,2)
	keys[0] = "blue3"
	newvals[0] = "blue7"
	newvals[1] = "some new thing"
	count, err := hs.Modify(3, "=", 1, 0, "U", keys, newvals)
	if err != nil {
		t.Error(err)
		}
	fmt.Println("modified", count, "records")




## Copyright and licensing

Licensed under **Apache License, version 2.0**.  
See file LICENSE.


## Contact

Brian Ketelsen - bketelsen@gmail.com

## Known bugs

No known bugs, but testing is far from comprehensive.

Working:  OpenIndex, Find, Insert,  Update/Delete


## Todo

Provide a layer of abstraction from the wire-level implementation of HandlerSocket to make a more intuitive interface.




## Credits and acknowledgments


Took some inspiration from the original GoMySQL implementation, although I've backed much of that out in this initial release.
https://github.com/Philio/GoMySQL
I can see how it would be extremely useful for GoMySQL or GoDBI to use HandlerSocket in the background for simple finds, inserts, etc.


## ChangeLog

1/20/2011
	Updated library extensively
	now working OpenIndex and Find commands
	
1/21/2011
	Insert works now
	
3/14/2011
	Modify and Delete work now - need more tests!


