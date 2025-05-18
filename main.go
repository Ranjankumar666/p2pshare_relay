package main

func main() {
	go CreateProxy()
	go CreateServer()
	select {}
}
