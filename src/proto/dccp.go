package proto
import (
    "strings"
    "io"
    "bytes"
    "os"
    "errors"
    "strconv"
    "fmt"
)

var HeaderLength int64 = 64

func ReadCommonHeader(src io.Reader)(string,error){
    buf :=make([]byte,2)
    _,err:=src.Read(buf)
    if err!=nil{
        return "",err
    }
    length,err:=strconv.Atoi(string(buf[:]))
    if err!=nil{
        return "",err
    }
    buf =make([]byte,length)
    _,err=src.Read(buf)
    if err!=nil{
        return "",err
    }
    header:=string(buf[:])
    if err!=nil{
        //log
        return "",err
    }
    return header,err
}

func WriteCommonHeader(writer io.Writer,header string)error{
    if int64(len(header))>HeaderLength{
        //log
        return errors.New("header too long")
    }
    var buf []byte = make([]byte,0)
    byteWriter := bytes.NewBuffer(buf)
    byteWriter.WriteString(fmt.Sprintf("%02d",len(header)))
    byteWriter.WriteString(header)
    _,err := byteWriter.WriteTo(writer)

    if err!=nil{

    }
    return err
}

func ReadFileToWriter(src io.Reader, des io.Writer)error{
    header,err := ReadCommonHeader(src)
    if err!=nil{
        return err
    }
    var size int64
    _,err=fmt.Sscanf(header,"type:file;size:%d",&size)
    if err !=nil{
        return err
    }
    _,err=io.CopyN(des,src,size)
    if err !=nil{
        return err
    }
    return nil
}

func ReadFileToFile(src io.Reader,des string)error{
    file,err := os.Create(des)
    if err!= nil{
        return err
    }
    defer file.Close()
    return ReadFileToWriter(src,file)
}

func ReadStringsSlice(src io.Reader)([]string,error){
    ret := make([]string,0)
    header,err := ReadCommonHeader(src)
    if err!=nil{

        return ret,err
    }

    sl := strings.Split(header,";")
    m := make(map[string]string)
    for _,v := range sl{
        s := strings.Split(v,":")
        m[s[0]]=s[1]
    }
    length,_ := strconv.Atoi(m["length"])
    _,_ = strconv.Atoi(m["count"])

    buf := new(bytes.Buffer)
    _,err = io.CopyN(buf,src,int64(length))
    if err!=nil{

    }else{

    }
    ret = strings.Split(buf.String(),"\n")

    return ret,nil
}

func ReadInt(src io.Reader)(int,error){
    header,err := ReadCommonHeader(src)
    if err != nil{
        return 0,err
    }
    var ret int
    sl := strings.Split(header,";")
    m := make(map[string]string)
    for _,v := range sl{
        s := strings.Split(v,":")
        m[s[0]]=s[1]
    }
    ret,err = strconv.Atoi(m["data"])
    return ret,err
}

func WriteString(des io.Writer,str string)error{
    header := fmt.Sprintf("type:string;size:%d\n",len(str))
    err := WriteCommonHeader(des,header)
    if err!=nil{

        return err
    }
    _,err=io.WriteString(des,str)
    if err !=nil{
        return err
    }
    return nil

}

func ReadString(src io.Reader)(string,error){
    header,err := ReadCommonHeader(src)
    if err !=nil{

        return "",err
    }
    var length int64

    sl := strings.Split(header,";")
    m:=make(map[string]string)
    for _,v := range sl{
        if len(v)>2{
            s2:=strings.Split(v,":")
            m[s2[0]]=s2[1]
        }
    }
    _,err = fmt.Sscanf(m["size"],"%d\n",&length)
    if err!=nil{

    }
    buf := new(bytes.Buffer)
    written,err := io.CopyN(buf,src,length)
    if written!=length{
        return "",errors.New("too less data")
    }
    if err!=nil{

    }
    str := buf.String()

    return str,nil

}

func WriteInt(des io.Writer,data int)error{
    header := fmt.Sprintf("type:int;data:%d",data)
    return WriteCommonHeader(des,header)
}

func WriteFileFromFile(file string,des io.Writer)(int64,error){
    fp,err := os.Open(file)
    if err!=nil{
        return 0,err
    }
    defer fp.Close()
    info,err :=fp.Stat()
    if err!=nil{
        return 0,err
    }
    return WriteFileFromReader(fp,des,info.Size())
}

func WriteFileFromReader(src io.Reader,des io.Writer,size int64)(int64,error){
    header := fmt.Sprintf("type:file;size:%d\n",size)
    err := WriteCommonHeader(des,header)
    if err != nil{
        return 0,err
    }
    return io.CopyN(des,src,size)
}

func WriteStringsSlice(des io.Writer,arr []string)error{
    tmp:=make([]string,len(arr)+1)
    copy(tmp,arr)
    tmp = append(tmp,"")
    join := strings.Join(arr,"\n")
    header := fmt.Sprintf("type:slice;count:%d;length:%d",len(arr),len(join))
    err := WriteCommonHeader(des,header)
    if err!= nil{
        return err
    }
    _,err=io.WriteString(des,join)
    if err!=nil{
        return err
    }
    return nil
}



