package main

import (
    "net"
    "proto"
    "fmt"
    "flag"
)

var balanceServer = flag.String("balance","localhost:2333","the balance server host:port")
var hostName = flag.String("host","localhost:2334","the remote host:port")
var status = flag.String("stat","avail","the stat of remote host:port")
var slots = flag.Int("slot",4,"the num slot of remote host:port")
var pindex = flag.Int("pindex",4,"the power index of remote host:port")

func main(){
    flag.Parse()
    balancer := *balanceServer
    hostname := *hostName
    stat := *status
    numSlots := *slots
    powerIndex := *pindex
    fmt.Println("got flag:",balancer,hostname,stat,numSlots,powerIndex)
    conn,err := net.Dial("tcp",balancer)
    if err!=nil{
        fmt.Println(err)
        return
    }
    dccp := proto.NewBalanceDccp(conn)
    err = dccp.SendStatus(hostname,stat,numSlots,powerIndex)
    if err!=nil{
        fmt.Println(err)
    }
}