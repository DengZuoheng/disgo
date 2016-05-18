package balancing

import (
    "net"
    "time"
    "os"
    "strings"
)


var SilentTime float64= 8


type Host struct {
    RawAddr string
    state State
    tcpAddr *net.TCPAddr
    nslots int
    pindex int
    lastUpdate time.Time
    ldAvg1,ldAvg5,ldAvg10 float64
}

func NewHost(addr string) *Host{
    host := &Host{}
    host.RawAddr=addr
    tcpaddr,err := net.ResolveTCPAddr("tcp",addr)
    if err==nil{
        host.tcpAddr=tcpaddr
    }else{
        host.tcpAddr=nil
    }
    host.state= StateAvailInstance()
    host.lastUpdate = time.Now()
    return host
}

func (self *Host)GetStatAsCode() StateCode{
    return self.state.AsCode()
}

func (self *Host)GetNumSlots() int{
    return self.nslots
}

func (self *Host)SetNumSlots(slots int){
    self.nslots=slots
}

func (self *Host)SetPindex(pindex int){
    self.pindex=pindex
}

func (self *Host)GetTier() int{

    return self.calcTier(self.ldAvg1,self.ldAvg5,self.ldAvg10,self.pindex)
}

func (self *Host)calcTier(ldAvg1,ldAvg5,ldAvg10 float64,pindex int) int{
    if ldAvg1 <0.9{
        return pindex
    }else if ldAvg5 <0.7{
        return pindex
    }else if ldAvg10 <0.8 {
        return pindex-1
    }else{
        return 0 //0表示不用这个host, 但是为毛这里返回0呢?
    }
}

func (self *Host)UpdateTier(ldAvg1,ldAvg5,ldAvg10 float64)(oldTier,newTier int){
    ldAvg1 /= float64(self.nslots)
    ldAvg5 /= float64(self.nslots)
    ldAvg10 /= float64(self.nslots)
    newTier = self.calcTier(ldAvg1,ldAvg5,ldAvg10,self.pindex)
    db := DBInstance()
    oldTier = self.GetTier()
    if newTier != oldTier{
        if newTier==0{
            self.Overloaded()
            self.ldAvg1=ldAvg1;self.ldAvg5=ldAvg5;self.ldAvg10=ldAvg10
        }else if oldTier==0{
            self.ldAvg1=ldAvg1;self.ldAvg5=ldAvg5;self.ldAvg10=ldAvg10
            self.Avail()//这里面会依赖tier, 所以先更新
        }else{
            db.MoveCpus(self,oldTier,newTier)
            self.ldAvg1=ldAvg1;self.ldAvg5=ldAvg5;self.ldAvg10=ldAvg10
        }
    }else{
        self.ldAvg1=ldAvg1;self.ldAvg5=ldAvg5;self.ldAvg10=ldAvg10
    }

    self.lastUpdate = time.Now()
    return
}

func (self *Host)Avail(){
    self.state.Avail(self)
}

func (self *Host)Unavail(){
    self.state.Unavail(self)
}

func (self *Host)Silent(){
    self.state.Silent(self)
}

func (self *Host)Overloaded(){
    self.state.Overloaded(self)
}

func (self *Host)ChangeState(state State){
    self.state = state
    self.state.AddToDB(self)//操, 这里又跟db耦合了
}

func (self *Host)SeemsDown()bool{
    delta := time.Since(self.lastUpdate) //since now
    return (delta.Seconds() > SilentTime)
}

func (self *Host)IsAvail() bool{
    return (self.state.AsCode() ==Available)
}

func (self *Host)IsUnavailable() bool {
    return (self.state.AsCode() ==Unavailable)
}

func (self *Host)IsSilent() bool {
    return (self.state.AsCode() ==Silent)
}

func (self *Host)IsOverloaded() bool {
    return (self.state.AsCode()  ==Overloaded)
}

//这两个函数也许不用写的
func (self *Host)String() string{
    return self.tcpAddr.String()
}

func (self *Host)Network() string{
    return self.tcpAddr.Network()
}

func (self *Host)IsLocal() bool{
    return false
    if strings.HasPrefix(self.RawAddr,"localhost"){
        return true
    }
    localIP := os.Getenv("DCC_LOCALADDR")
    if localIP!=""{
        return strings.HasPrefix(self.RawAddr,localIP)
    }
    return false
    //ref: http://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
}

func (self *Host)Connect() (net.Conn,error){
    return net.DialTimeout("tcp",self.tcpAddr.String(),60*time.Second)
}