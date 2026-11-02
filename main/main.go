package main

import (
	"fmt"
	"fakery"
)

func main() {
	
	f := fakery.New();
	a := f.Address();
	fmt.Println(a);

}
