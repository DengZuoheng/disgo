package main

import (
    "balancing"
    "runtime"
)

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU()+2)
    ch := make(chan bool,1)
    go balancing.ListenRequest("2333")
    _ :<-ch
}
