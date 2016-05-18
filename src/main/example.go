package main

import (
    "strings"
    "fmt"
)

func ExpandPreprocessorOptions(args []string)[]string{
    ret := make([]string,0)
    for i:= 0;i<len(args);i++{
        if strings.HasPrefix(args[i],"-Wp,"){
            extras := strings.Split(args[i],",")
            for j:=1;j<len(extras);j++{
                ret=append(ret,extras[j])
                if (extras[j]=="-MD"||extras[j]=="-MMD") && j<len(extras){
                    ret=append(ret,"-MF")
                    ret=append(ret,extras[j+1])
                }
            }
        }else{
            ret=append(ret,args[i])
        }
    }
    return ret

}

func main(){
    argv := make([]string,0)
    argv = append(argv,"gcc")
     argv = append(argv,"-c")
     argv = append(argv,"fuck.c")
     argv = append(argv,"-ifukck")
     argv = append(argv,"-Ishit")
     argv = append(argv,"-Wp,-MD,-MF,-MDD")
     argv = append(argv,"-std=c99")
     argv = append(argv,"-j")
    argv=ExpandPreprocessorOptions(argv)
    fmt.Println(argv)


}
