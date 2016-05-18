package compile
import(
    "balancing"
    "proto"
    "net"
    "os"
    "time"

)

const(
    Succeed=iota
)

type LocalCompileController interface {
    BuildLocal()(status int, err error)
    BuildRemote(hostCh chan *balancing.Host,backCh chan string)(status int, err error)
    MustLocal() bool
    PrepareForCompile() (status int,err error)
}

type LocalCompileControllerBase struct{
    CompileName string
    Args []string
}

type LocalCompileControllerOuter struct{
    ctrl LocalCompileController
}

func (self *LocalCompileControllerOuter) BuildSomewhere() (status int,err error){
    if self.ctrl.MustLocal(){//也许sgLelel就决定了必须本地了
        //go WarnLdAvgSrvForLoal(chw)//也许需要通知一下负载均衡器

        return self.ctrl.BuildLocal()
    }

    out := make(chan *balancing.Host,1)
    back := make(chan string,1)
    go GetHost(out,back)//这里可能未归还就已经进程完结了，所以要阻塞等待一下

    status,err = self.ctrl.PrepareForCompile()
    if err!=nil{
        //log
        _,ok:=<-out
        if ok{
            back<-"local"
            _,ok =<-out
        }

        return self.ctrl.BuildLocal()

    }
    if self.ctrl.MustLocal(){
        _,ok:=<-out
        if ok{
            back<-"local"
            _,ok =<-out
        }
        return self.ctrl.BuildLocal()
    }

    status,err =  self.ctrl.BuildRemote(out,back)
    _,ok:=<-out
    if ok{
        //
    }
    return
}

func (self *LocalCompileControllerOuter) BuildSomewhereTimed() (status int,err error){

    //gettime
    status,err = self.BuildSomewhere()
    if err!=nil{

    }
    //gettime
    //log timedelta
    //log err
    return status,err
}

func GetHost(out chan *balancing.Host,back chan string){
    defer close(out)
    server := os.Getenv("DCCLB_SERVER")
    if server==""{
        server="localhost"
    }
    port := os.Getenv("DCCLB_PORT")
    if port==""{
        port="2333"
    }

    addr := net.JoinHostPort(server,port)

    conn, err := net.DialTimeout("tcp",addr,8*time.Second)
    if err != nil{

        return
    }
    dccp := proto.NewBalanceDccp(conn)
    err = dccp.SendHostRequest()
    if err != nil {

        return
    }
    host,err := dccp.RetrieveHost()
    if err!= nil{

        return
    }
    out<-balancing.NewHost(host)
    select{
    case status :=<- back:
        dccp.ReleaseHost(status)
    case <-time.After(60*time.Second)://超时机制

        dccp.ReleaseHost("timeout")
    }
    out<-nil
}

func MakeLocalCompileController(sgLevel int, compilerName string, args []string) *LocalCompileControllerOuter{
    var ctrl LocalCompileController
    switch compilerName {
    case "gcc":
        fallthrough
    case "g++":
        ctrl= NewGCCLocalCompileController(sgLevel, args)
    default:
        ctrl= NewGCCLocalCompileController(sgLevel, args)

    }
    return &LocalCompileControllerOuter{ctrl:ctrl}

}








