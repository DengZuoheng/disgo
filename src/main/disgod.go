package main

import (
    "daemon"
    "os"
    "net"
    "strconv"
    "fmt"
    "runtime"
    "flag"
)

var balanceServer = flag.String("balance","","the balance server host:port")
var hostName = flag.String("host","","the remote host:port")
var portListen = flag.String("port","2334","")

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU()+2)
    flag.Parse()
    ch := make(chan bool,1)
    addr := os.Getenv("LDAVG_SERVER")
    port := os.Getenv("DCCD_PORT")
    if addr=="" {
        addr= *balanceServer
        fmt.Println(*balanceServer)
    }
    localName := *hostName
    fmt.Println("get ldsrv addr:",addr)
    fmt.Println("get local name:",localName)
    tcpaddr,err := net.ResolveTCPAddr("tcp",addr)
    if err!=nil{
       fmt.Println("tcpaddr resolve tcp addr fail:",tcpaddr)
    }
    if port==""{
        port=*portListen
    }
    localSrvPort,err := strconv.Atoi(port)
    if err!=nil{
        fmt.Println("port atoi fail:",err)
    }
    go daemon.DaemonSrv(*tcpaddr,localSrvPort,localName)
    _=<-ch //永远阻塞了
}