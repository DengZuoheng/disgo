package balancing
import (
    "net"
    "proto"
    "time"
    "fmt"
)

//读取msg, 丢到队列
func HandleConn(conn net.Conn, msgQueueCh chan Msg){
    dccp := proto.NewBalanceDccp(conn)
    msg,err := dccp.ReadLdMsg()
    if err != nil{
        fmt.Println("read ld msg err from:",conn.RemoteAddr().String(),err,"going to treat as release")
        //出错的多半是释放的
        msg=make(map[string]string,0)
        msg["type"]="release"
    }
    msgObj := MakeMsg(conn,msgQueueCh,msg)
    if msgObj!=nil{
        msgQueueCh <- msgObj
    }else{
        conn.Close()
    }

}

func ResponseHostReq(dccp *proto.BalancerSideDccp,addr string){

    dccp.WriteHostResponse(addr)
}

func LoadBalancingServer(){
    go DoSlientSearch()
    port := "2333"
    go ListenRequest(port)
}


func ListenRequest(port string){
    ln,err := net.Listen("tcp",fmt.Sprintf(":%s",port))
    if err != nil{
        //log err
        fmt.Println("listen err:",err)
        return
    }
    fmt.Println("listening prot",port)
    msgQueueCh := make(chan Msg,1000)//给conn排队, 主go程select到就往里面写, 然后继续开协程

    go SerializeHandle(msgQueueCh)
    for{

        conn, err := ln.Accept()
        if err != nil{
            fmt.Println("accept err:",err)
            //log即使我知道出错我也没办法啊
        }
        //conn本身可以设置读写超时, 应该在handle中根据消息类型设置尝试
        fmt.Println("accepted:",conn.RemoteAddr().String())
        go HandleConn(conn,msgQueueCh)
    }

}

//消息队列, 一个一个消息处理
func SerializeHandle(msgQueueCh chan Msg){
    //将并行收束成串行
    //实际上不需要select, 直接接收Msg指针然后handle就好
    for{
        msg,ok := <-msgQueueCh
        if ok {
            msg.Handle()
        }else{
            break
        }
    }
}

func DoSlientSearch() {
    for{
        DBInstance().HandleSilentHosts()

        time.Sleep(60*time.Second)

    }
}

