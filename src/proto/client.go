package proto

import (
    "net"
    "io"
)

type ClientSideDccp struct{
    Conn net.Conn
}

func NewClientDccp(conn net.Conn) *ClientSideDccp{
    return &ClientSideDccp{Conn:conn}
}

func (self *ClientSideDccp)SendHeader() error{
    return WriteString(self.Conn,"type:compile;cctype:gcc;cpp:local")

}

func (self *ClientSideDccp)SendArgs(args []string) error{
    return WriteStringsSlice(self.Conn,args)
}

func (self *ClientSideDccp)NoticeFailAtCpp() error{
    str := "fail:cpp"
    return WriteString(self.Conn,str)
}

func (self *ClientSideDccp)SendSourceFile(filename string) error{
    _,err :=  WriteFileFromFile(filename,self.Conn)
    return err
}

func (self *ClientSideDccp)RetrieveStatus()(int,error){
    return ReadInt(self.Conn)
}

func (self *ClientSideDccp)RetrieveOutputByFile(outputFname string)error{
    return ReadFileToFile(self.Conn,outputFname)
}

func (self *ClientSideDccp)RetrieveStderrByWriter(writer io.Writer)error{
    return ReadFileToWriter(self.Conn,writer)
}

func (self *ClientSideDccp)RetrieveStdoutByWriter(writer io.Writer)error{
    return ReadFileToWriter(self.Conn,writer)
}