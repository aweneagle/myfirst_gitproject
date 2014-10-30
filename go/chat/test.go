package main

const A = 1
const B = 2
const C = A + B

func main() {
	defer func () {
		recover()
	}()
	panic(C)
}

