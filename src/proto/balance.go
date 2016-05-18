package proto

import (
    "net"
    "strings"
    "fmt"
    "errors"
)
type BalancerSideDccp struct{
    Conn net.Conn
}

func NewBalanceDccp(conn net.Conn)*BalancerSideDccp{
    return &BalancerSideDccp{Conn:conn}
}

func (self *BalancerSideDccp) ReadLdMsg()(map[string]string,error){
    str,err := ReadString(self.Conn)
    if err!=nil{
        return nil,err
    }

    if strings.HasPrefix(str,"type:host"){
        var cliIpStr string
        _,err := fmt.Sscanf(str,"type:host %s",&cliIpStr)
        if err==nil{
            msg:=make(map[string]string,2)
            msg["type"]="host"
            msg["addr"]=cliIpStr

            return msg,nil
        }
        return nil,err

    }else if strings.HasPrefix(str,"type:load"){
        var ldavg1,ldavg5,ldavg10 string
        var machname  string
        _,err := fmt.Sscanf(str,
            "type:load %s %s %s %s",
            &machname,&ldavg1,&ldavg5,&ldavg10 )
        if err==nil{
            msg := make(map[string]string,5)
            msg["type"]="load"
            msg["addr"]=machname
            msg["ldavg1"]=ldavg1
            msg["ldavg5"]=ldavg5
            msg["ldavg10"]=ldavg10
            return msg,nil//这个再说吧
        }
    }else if strings.HasPrefix(str,"type:status"){
        var machname,state string
        var numslots  string
        var pindex string
        _,err := fmt.Sscanf(str,
            "type:status %s %s %s %s",
            &machname, &state,&numslots,&pindex)
        if err == nil{
            msg := make(map[string]string,5)
            msg["type"]="status"
            msg["addr"]=machname
            msg["state"]=state
            msg["numslots"]=numslots
            msg["pindex"]=pindex
            return msg,nil
        }
    }else if strings.HasPrefix(str,"type:monitor"){
        msg := make(map[string]string,1)
        msg["type"]="monitor"
        return msg,nil
    }else if strings.HasPrefix(str,"type:release"){
        msg := make(map[string]string,1)
        msg["type"]="release"
        return msg,nil
    }

    return nil,err
}


func (self *BalancerSideDccp)SendHostRequest()error{
    reqstr := fmt.Sprintf("type:host %s",self.Conn.LocalAddr().String())
    return WriteString(self.Conn,reqstr)
}

func (self *BalancerSideDccp)RetrieveHost()(string,error){
    repstr,err := ReadString(self.Conn)
    if err!=nil{
        return "",err
    }
    if strings.HasPrefix(repstr,"type:hostret"){
        var addr string
        _,err := fmt.Sscanf(repstr,"type:hostret %s",&addr)
        return addr,err

    }
    return "",errors.New("unexpect addr")
}

func (self *BalancerSideDccp)WriteHostResponse(addr string)error{
    repstr := fmt.Sprintf("type:hostret %s",addr )
    return WriteString(self.Conn,repstr)
}

func (self *BalancerSideDccp)ReleaseHost(status string)error{
    repstr := fmt.Sprintf("type:release %s",status)
    return WriteString(self.Conn,repstr)
}

func (self *BalancerSideDccp)SendStatus(addr,status string,slots,pindex int)error{
    stat := fmt.Sprintf("type:status %s %s %d %d",addr,status,slots,pindex)
    return WriteString(self.Conn,stat)
}