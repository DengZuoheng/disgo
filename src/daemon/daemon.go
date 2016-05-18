package daemon

import (
    "net"
    "time"
    "proto"
    "fmt"
    "runtime"
    "os"
    "strconv"
    "compile"
    "syscall"
)

var LINUX_SYSINFO_LOADS_SCALE float32=65536.0

func ReportLdAvg(localName string,addr net.TCPAddr){
    for{
        beforeDial := time.Now()
        conn,err := net.DialTimeout("tcp",addr.String(),10*time.Second)
        if err==nil{
            var LdAvg1,LdAvg5,LdAvg10 float32
            var info syscall.Sysinfo_t
            err := syscall.Sysinfo(&info)
            if err!=nil{

                time.Sleep(10)
                conn.Close()
                continue
            }
            LdAvg1 = float32(info.Loads[0])/LINUX_SYSINFO_LOADS_SCALE
            LdAvg5 = float32(info.Loads[1])/LINUX_SYSINFO_LOADS_SCALE
            LdAvg10 = float32(info.Loads[2])/LINUX_SYSINFO_LOADS_SCALE
            dccp := proto.NewDaemonDccp(conn)
            dccp.ReportLdAvg(localName,LdAvg1,LdAvg5,LdAvg10)
            fmt.Println(time.Now(),"reported ldavg:",LdAvg1,LdAvg5,LdAvg10)
            conn.Close()
        }else{
            fmt.Println("disgob dial fail:",err)
        }
        delta := time.Since(beforeDial)
        if delta < 10*time.Second{//10秒钟一次报告
            time.Sleep(10*time.Second)
        }

    }
}

func ConcurrencyLevel() int{
    var slots int
    limit := os.Getenv("DCC_CONCURRENCY_LIM")
    if limit!=""{
        i,err := strconv.Atoi(limit)
        if err ==nil{
            slots=i
        }
    }else{
        slots = runtime.NumCPU()+2

    }
    return slots
}



func ServiceJobToken(conn net.Conn,ch chan<- bool){
    fmt.Println("now we going to service job from:",conn.RemoteAddr().String())
    dccp := proto.NewDaemonDccp(conn)
    cctype,features,ok := dccp.ReadHeader()
    if !ok{
        fmt.Println("read header not ok")
        //fail
        //return
    }
    cchdlr := compile.MakeRemoteCompileHandler(cctype,features,dccp)
    ok =cchdlr.PrepareForCompile()
    if ok{
        cmd := cchdlr.Cmd()
        if cmd!=nil{
            cmd.Start()
            cmd.Wait()
        }
        ch<-true //交还令牌, 异步回复
        go func(hdlr compile.RemoteCompileHandler){
            hdlr.Response()
        }(cchdlr)
    }else {
        ret :=cchdlr.Response()
        if ret{
            fmt.Println("response to",conn.RemoteAddr().String(),"all finished!")
        }else{

        }
        conn.Close()
        ch <- true
    }
}

func ServiceJob(conn net.Conn){
    dccp := proto.NewDaemonDccp(conn)
    cctype,features,ok := dccp.ReadHeader()
    if !ok{

        //fail
        //return
    }
    cchdlr := compile.MakeRemoteCompileHandler(cctype,features,dccp)
    ok =cchdlr.PrepareForCompile()
    if ok{
        cmd := cchdlr.Cmd()
        cmd.Start()
        cmd.Wait() //其实这里也是异步回复
        //压缩和加密都是消耗cpu的, 所以其实没必要异步
        go func(hdlr compile.RemoteCompileHandler){
            hdlr.Response()
        }(cchdlr)
    }

}


func ConcurrencyPatternPreSpawn(listener net.Listener,clevel int){
    for i:=0;i<clevel;i++{
        go func(l net.Listener){
            for{
                conn,err := l.Accept()
                if err!= nil{
                    ServiceJob(conn)
                }
            }
        }(listener)
    }
}

func ConcurrencyPatternToken(listener net.Listener, clevel int){

    ch := make(chan bool,clevel)
    for i:=0;i<clevel;i++{
        ch<-true
    }
    for{
        _,ok := <-ch
        if ok{
            conn,err := listener.Accept()
            if err != nil{
                //log
                fmt.Println("accept err:",err)
                ch<-true
            }
            fmt.Println("compile srv accept a job from:",conn.RemoteAddr().String())
            go ServiceJobToken(conn,ch)
        }
    }
}

func RemoteCompileListen(port int){
    l,err := net.Listen("tcp",fmt.Sprintf(":%d",port))
    if err != nil{
        //log
        fmt.Println("listen fail:",err)
        return
    }
    c := ConcurrencyLevel()
    fmt.Println("got concurrency level:",c)
    fmt.Println("listening:",port)
    ConcurrencyPatternToken(l,c)
    //ConcurrencyPatternPreSpawn(l.c)

}


func DaemonSrv(balancingServerAddr net.TCPAddr,localSrvPort int,localName string){
    go ReportLdAvg(localName,balancingServerAddr)
    go RemoteCompileListen(localSrvPort)
}

