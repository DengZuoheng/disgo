package balancing

import(
    "net"

    "proto"
    "strconv"
    "fmt"
)

type Msg interface {
    Handle() error
}

type MsgBase struct{
    Conn net.Conn
    MsgQueueCh chan Msg
    Addr string
    Dprop string
}

func (self *MsgBase) Handle(){
    //pass
}

type HostReqMsg struct {
    MsgBase
    dccp *proto.BalancerSideDccp
}

type StatMsg struct {
    MsgBase
    NumSlots int
    PowerIndex int
    Status StateCode
    RawAddr string
}

type MonReqMsg struct{
    MsgBase
}

type LdAvgMsg struct{
    MsgBase
    LdAvg1,LdAvg5,LdAvg10 float64

}

type ReleaseMsg struct{
    MsgBase
}

func MakeMsg(conn net.Conn, msgQueueCh chan Msg,data map[string]string) Msg{
    if data["type"]=="host"{
        msg := &HostReqMsg{}
        msg.Addr=data["addr"]
        msg.Conn=conn
        msg.dccp=proto.NewBalanceDccp(conn)
        msg.MsgQueueCh=msgQueueCh
        return msg
    }else if data["type"]=="load"{
        msg := &LdAvgMsg{}
        msg.Addr=data["addr"]
        msg.LdAvg1,_ = strconv.ParseFloat(data["ldavg1"],32)
        msg.LdAvg5,_ = strconv.ParseFloat(data["ldavg5"],32)
        msg.LdAvg10,_ = strconv.ParseFloat(data["ldavg10"],32)
        msg.Conn=conn
        msg.MsgQueueCh=msgQueueCh
        return msg
    }else if data["type"]=="status"{
        msg := &StatMsg{}
        if data["state"]=="avail"{
            msg.Status=Available
        }else if data["state"]=="unavail"{
            msg.Status=Unavailable
        }
        msg.Addr = data["addr"]
        msg.RawAddr = data["addr"]
        msg.NumSlots,_ = strconv.Atoi(data["numslots"])
        msg.PowerIndex,_ = strconv.Atoi(data["pindex"])
        msg.Conn=conn
        msg.MsgQueueCh=msgQueueCh
        return msg
    }else if data["type"]=="monitor"{
        msg := &MonReqMsg{}
        return msg
    }else if data["type"]=="release"{
        msg := &ReleaseMsg{}
        msg.Conn=conn
        return msg
    }
    return nil

}


func (self *ReleaseMsg) Handle()error{
    defer self.Conn.Close()
    db := DBInstance()
    db.ReleaseSlot(self.Conn)
    fmt.Println("release slot from:",self.Conn.RemoteAddr().String())
    return nil
}


//响应host请求
func (self *HostReqMsg) Handle()error{
    //这里先不关闭连接, 因为客户端拿到host后可能还会
    db := DBInstance()
    addr,ok := db.GetBestAvailSlot()
    if !ok {
        addr = "localhost"
        //把它自己给他
    }else{
         //log
        db.AssignCpuToClient(addr,self.Conn)
    }
    go ResponseHostReq(self.dccp,addr)
    go HandleConn(self.Conn,self.MsgQueueCh)
    fmt.Printf("host request from %s ret by %s\n",self.Conn.RemoteAddr().String(),addr)
    return nil
}

//设置host可用或不可用
func (self *StatMsg) Handle()error{

    if self.Conn!=nil{
        defer self.Conn.Close()
    }
    db := DBInstance()
    if(self.Status == Available){

        if db.HaveHost(self.Addr){
            host,ok := db.GetHost(self.Addr)
            if ok{
                host.Avail()
            }
        }else{

            host := NewHost(self.Addr)
            host.SetNumSlots(self.NumSlots)
            host.SetPindex(self.PowerIndex)
            db.AddNewHost(host)

        }
    }else{
        host,ok := db.GetHost(self.Addr)
        if ok {
            host.Unavail()
        }

    }
    fmt.Println("stat msg from:",self.Conn.RemoteAddr().String())
    return nil
}

func (self *MonReqMsg) Handle()error{
    //pass
    //这是管理信息, 以后再说
    defer self.Conn.Close()
    return nil
}

func (self *LdAvgMsg) Handle()error{

    if self.Conn!=nil{
         defer self.Conn.Close()
    }

    db := DBInstance()

    host,ok := db.GetHost(self.Addr)
    if ok{

        host.UpdateTier(self.LdAvg1,self.LdAvg5,self.LdAvg10)
        fmt.Println("ld avg from:",self.Addr,self.LdAvg1,self.LdAvg5,self.LdAvg10)

    }else{

        return nil

    }

    return nil
}
