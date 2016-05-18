package main

import (
	"os"
    "path"
    "safeguard"
    "compile"
    "runtime"
)



func Fail(err error){

}

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU()+2)
	args := os.Args
	compilerName := path.Base(args[1])
    //util.args.ExpandArgs(args) //这个以后在说
    var sgLevel = safeguard.GetRecursionLevel()
	ctrl := compile.MakeLocalCompileController(sgLevel,compilerName,args[1:])
    //compile需要设置一些开关, 以启停pipeline
    ret,err := ctrl.BuildSomewhereTimed()
    if err == nil {
        os.Exit(0)
    }else{
        os.Exit(ret)
    }

}



