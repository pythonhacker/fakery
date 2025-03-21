package main

import (
	"fmt"
	"gofakelib"
)

func main() {

	//	f := gofakelib.NewFromSeed(1)
	f := gofakelib.New()

	fmt.Println(f.IntRange(1000))
	per := f.Person()
	fmt.Printf("%s\n", per)
	fmt.Println(f.Email())
	fmt.Println(f.Name())
	fmt.Println(f.City())
	fmt.Println(f.StreetAddress())
	fmt.Println(f.State())
	fmt.Println(f.PostCode())
	fmt.Println("*********")
	fmt.Println(f.Address())
	fmt.Println(f.Beer())
	fmt.Println(f.Wine())
	//	f.SetLocale("en_GB")

	//	fmt.Println(f.Name())
}
