package proto

import (
    "net"
    "strings"
    "fmt"
)

type DaemonSideDccp struct{
    Conn net.Conn
}

func NewDaemonDccp(conn net.Conn) *DaemonSideDccp{
    return &DaemonSideDccp{Conn:conn}
}

func (self *DaemonSideDccp)ReadHeader()(string,map[string]string,bool){
    header,err :=ReadString(self.Conn)
    if err!=nil{
        fmt.Println("read header fail:",err)
    }

    var cctype string
    sl := strings.Split(header,";")
    m := make(map[string]string,0)
    for _,v := range sl{
        s := strings.Split(v,":")
        m[s[0]]=s[1]
    }
    cctype = m["cctype"]
    return cctype,m,true

}


func (self *DaemonSideDccp)ReadArgs()([]string,error){
    return ReadStringsSlice(self.Conn)
}

func (self *DaemonSideDccp)ReadSourceFile(filename string)error{
    return ReadFileToFile(self.Conn,filename)
}

func (self *DaemonSideDccp)SendStatus(status int)error{
    return WriteInt(self.Conn,status)
}

func (self *DaemonSideDccp)SendObjectFile(filename string)error{
    _,err := WriteFileFromFile(filename,self.Conn)
    return err
}

func (self *DaemonSideDccp)SendStdErrFile(filename string)error{
    _,err := WriteFileFromFile(filename,self.Conn)
    return err
}

func (self *DaemonSideDccp)SendStdOutFile(filename string)error{
    _,err := WriteFileFromFile(filename,self.Conn)
    return err
}

func (self *DaemonSideDccp)ReportLdAvg(LocalName string,LdAvg1,LdAvg5,LdAvg10 float32)error{
    str := fmt.Sprintf("type:load %s %f %f %f",LocalName,LdAvg1,LdAvg5,LdAvg10)
    return WriteString(self.Conn,str)
}
