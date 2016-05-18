package tmp

import (
    "math/rand"
    "time"
    "os"
    "fmt"
    "path"
)


func init() {
    rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    l := len(letterRunes)
    for i := range b {
        b[i] = letterRunes[rand.Intn(l)]
    }
    return string(b)
}

func TempTopDir() string{
    d := os.Getenv("DCC_TMPDIR")
    if d!=""{
        return d
    }
    return os.TempDir()
}

func TempTopDirPerm() (string,os.FileMode,error){
    tempdir := TempTopDir()
    fileinfo,err := os.Stat(tempdir)
    var perm os.FileMode
    if err != nil{
        //log
        return "",perm,err
    }
    perm = fileinfo.Mode().Perm()
    return tempdir,perm,nil
}

func MakeTempName(prefix,suffix string)(tmpfpath string,err error){
    tempdir,perm,err := TempTopDirPerm()
    if err!=nil{
        return
    }
    //这里需要判断一下access perm
    for{
        randstr := RandStringRunes(8)

        tmpfname := fmt.Sprintf("%s_%s%s",prefix,randstr,suffix)
        tmpfpath := path.Join(tempdir,tmpfname)
        //然后去open
        f,err:= os.OpenFile(tmpfpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL,perm)
        if err != nil{
            //log
            continue
        }
        f.Close()
        //add to cleanup
        return tmpfpath,nil
    }
    return

}

func RemoveIfExists(fname string){
    os.Remove(fname)
}

func NewTempDir()(string,error){
    tempdir,perm,err := TempTopDirPerm()
    if err != nil{
        return "",err
    }
    randstr := RandStringRunes(8)
    tmpdirname := fmt.Sprintf("distccd_%s",randstr)
    tmpdirpath := path.Join(tempdir,tmpdirname)
    err = os.MkdirAll(tmpdirpath,perm)
    if err != nil{
        //log
        return "",err
    }
    return tmpdirpath,nil
}

func MakeServerCwd(tmpdir string,clientcwd string)(string,error){
    _,perm,err := TempTopDirPerm()
    if err != nil{
        return "",err
    }
    servercwd := fmt.Sprintf("%s%s",tmpdir,clientcwd)
    err = os.MkdirAll(servercwd,perm)
    return servercwd,err
}
