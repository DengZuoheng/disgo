package gcc

import (
    "strings"
    "path/filepath"
    "path"
    "fmt"
)

//如果true, 说明这种文件是需要预处理的
var SourceFileExt = map[string] bool {
    ".i":false,
    ".ii":false,
    ".c":true,
    ".cc":true,
    ".cpp":true,
    ".cxx":true,
    ".cp":true,
    ".c++":true,
    ".C":true,
    ".m":true,
    ".mm":true,
    ".mi":false,
    ".mii":false,
    ".M":true,
    ".s":false,
    ".S":true,
}

var IncludeOptPrefix = []string {
    "-I",
    "-include",
    "-imacros",
    "-idirafter",
    "-iprefix",
    "-iwithprefix",
    "-iwithprefixbefore",
    "-isystem",
    "-iquote",
}

var PreprocExtMap = map[string]string{
    ".i":".i",
    ".c":".i",
    ".c++":".ii",
    ".cc":".ii",
    ".cpp":".ii",
    ".cxx":".ii",
    ".cp":".ii",
    ".C":".ii",
    ".ii":".ii",
    ".mi":".mi",
    ".m":".mi",
    ".mii":".mii",
    ".mm":".mii",
    ".M":".mii",
    ".s":".s",
}


func OptMustLocal(opt string)bool {
    if opt[1]=='M'{
        return true
    }else if opt=="-march=native"{
        return true
    }else if opt=="-mtune=native"{
        return true
    }else if strings.HasPrefix(opt,"-Wa,"){
        if strings.Contains(opt,",-a") || strings.Contains(opt,"--MD"){
            return true
        }
    }else if strings.HasPrefix(opt,"-specs="){
        return true
    }else if opt =="-fprofile-arcs"||opt=="-ftest-coverage"||opt=="--coverage"{
        return true
    }else if opt =="-frepo"{
        return true
    }else if strings.HasPrefix(opt,"-x"){
        return true
    }else if strings.HasPrefix(opt,"-dr"){
        return true
    }else{
        return false
    }
    return true
}


func SetOutput(args []string,filename string)string{
    for i:=0;i<len(args);i++{
        if args[i]=="-o" && i!=len(args)-1{
            ret := args[i+1]
            args[i+1]=filename
            return ret
        }else if strings.HasPrefix(args[i],"-o"){
            ret := args[i+1]
            args[i]=fmt.Sprintf("-o%s",filename)
            return ret
        }
    }
    //如果找不到, 就append一个
    args=append(args,"-o")
    args=append(args,filename)
    return ""
}


func SetInput(args []string,filename string)string{
    for i:=0;i<len(args);i++{
        if IsSource(args[i]){
            ret := args[i]
            args[i]=filename
            return ret
        }
    }
    //如果找不到, 就不管
    return ""
}

func SourceNeedsLocal(filename string)bool{
    base := filepath.Base(filename)
    if strings.HasPrefix(base,"conftest.") || strings.HasPrefix(base,"tmp.conftest."){
        return true
    }
    return false
}

func IsSource(filename string) bool{
    ext := filepath.Ext(filename)
    _,ok := SourceFileExt[ext]
    return ok
}

func AppendOutputFromSource(args []string,source,ext string) (string,[]string){
    base := filepath.Base(source)
    dot := strings.LastIndex(base,".")
    output := fmt.Sprintf("%s%s",base[:dot],ext)
    args=append(args,"-o")
    args=append(args,output)
    return output,args

}

func ScanArgs(args []string)(argv []string,inputfile,outputfile string,ok bool){
    ok=false
    seenOptS:=false
    seenOptc:=false
    argv = args
    outputfile=""
    if len(argv) == 0{
        return
    }
    if len(argv[0])==0{
        return
    }
    if argv[0][0] == '-'{
        return
    }
    i :=0
    for{
        a:=argv[i]
        if a[0]=='-'{
            if a=="-E"{
                //must local
                return
            }else if a=="-MD" || a=="-MMD"{

            }else if a=="-MG" || a=="-MP"{

            }else if a=="-MF" || a=="-MT" || a=="-MQ"{
                i++
            }else if (strings.HasPrefix(a,"-MF") ||
                strings.HasPrefix(a,"-MT") || strings.HasPrefix(a,"-MQ")){

            }else if OptMustLocal(a){
                //must local
                return
            }else if a=="-S"{
                seenOptS=true
            }else if a=="-c"{
                seenOptc=true
            }else if strings.HasPrefix(a,"-o"){
                if len(a)==2{
                    outputfile=argv[i+1]
                    i++
                }else{
                    outputfile=argv[i][2:]
                }
            }
        }else{
            if IsSource(a){
                inputfile=a
            }else if strings.HasSuffix(a,".o"){
                outputfile=a
            }
        }
        i++
        if i>=len(argv){
            break
        }
    }
    if !(seenOptS||seenOptc){
        return
    }

    if SourceNeedsLocal(inputfile){
        return
    }

    if outputfile==""{
        if seenOptc{
            outputfile,argv = AppendOutputFromSource(argv,inputfile,".o")
        }else if seenOptS{
            outputfile,argv = AppendOutputFromSource(argv,inputfile,".s")
        }

    }

    if outputfile=="-"{
        return
    }
    ok = true
    return

}

//dcc_convert_mt_to_dotd_target
func PopMtOption(args []string)(string,bool){
    for i:=0;i<len(args);i++{
        if args[i]=="-MT"{
            if i == len(args)-1{
                //fail
                return "",false
            }
            ret := args[i]
            args = append(args[:i],args[i+1:]...)
            return ret,true
        }
    }
    return "",false
}

func TweakIncludesArgForServer(args []string,rootdir string)int{
    count :=0
    for i:=0;i<len(args);i++{
        for _,v := range IncludeOptPrefix{
            if args[i] == v && i!= len(args)-1{
                args[i+1]=filepath.Join(rootdir,args[i+1])
                count++
            }else if strings.HasPrefix(args[i],v){
                args[i]=fmt.Sprintf("%s%s",v,filepath.Join(rootdir,args[i][len(v):]))
                count++
            }
        }
    }
    return count
}

func TweakInputArgForServer(args []string,rootdir string)string{
    for i:=0;i<len(args);i++{
        if IsSource(args[i]) &&path.IsAbs(args[i]){
            ret:=args[i]
            args[i]=path.Join(rootdir,args[i])
            return ret
        }
    }
    return ""
}

func TweakArgsForServer(args []string, rootdir,depfname string)(string,bool){
    var dotdTarget string=""
    var ok bool=false
    if dotdTarget,ok = PopMtOption(args);!ok{
        return dotdTarget,ok
    }
    args=append(args,"-MMD")
    args=append(args,"-MF")
    args=append(args,depfname)
    TweakIncludesArgForServer(args,rootdir)
    TweakInputArgForServer(args,rootdir)

    return dotdTarget,ok
}

func PreprocExten(sourceExt string)string{
    v,ok := PreprocExtMap[sourceExt]
    if ok{
        return v
    }
    return ""
}

func IsPreprocessed(filename string)bool{
    ext:=filepath.Ext(filename)
    needCPP,ok := SourceFileExt[ext]
    if ok{
        return !needCPP
    }
    return false//其实false意味着我也不知道
}

func StripDasho(args []string) ([]string,int){
    count := 0
    for i:=0;i<len(args);i++{
        if strings.HasPrefix(args[i],"-o"){
            if len(args[i])>len("-o"){ //-ofname 跳过一个
                args = append(args[:i],args[i+1:]...)
                count++
            }else{ //-o fname 跳过两个
                args = append(args[:i],args[i+2:]...)
                count++
            }
        }
    }
    return args,count
}

func SetActionOpt(args []string,opt string) bool{
    for i:=0;i<len(args);i++{
        if args[i]=="-c"||args[i]=="-S"{
            args[i]=opt
            //log
            return true
        }
    }
    return false
}


func StripLocalArgs(args []string) []string{
    ret := make([]string,0)//反正内存便宜
    for i:=0;i<len(args);i++{
        if (args[i]=="-D"||
        args[i]=="-I"||
        args[i]=="-U"||
        args[i]=="-L"||
        args[i]=="-l"||
        args[i]=="-MF"||
        args[i]=="-MT"||
        args[i]=="-MQ"||
        args[i]=="-include"||
        args[i]=="-imacros"||
        args[i]=="-iprefix"||
        args[i]=="-iwithprefix"||
        args[i]=="-isystem"||
        args[i]=="-iwithprefixbefore"||
        args[i]=="-idirafter"){
            if i<len(args)-1{
                i++
            }
        }else if(strings.HasPrefix(args[i],"-Wp,")||
        strings.HasPrefix(args[i],"-Wl,")||
        strings.HasPrefix(args[i],"-D")||
        strings.HasPrefix(args[i],"-U")||
        strings.HasPrefix(args[i],"-I")||
        strings.HasPrefix(args[i],"-l")||
        strings.HasPrefix(args[i],"-L")||
        strings.HasPrefix(args[i],"-MF")||
        strings.HasPrefix(args[i],"-MT")||
        strings.HasPrefix(args[i],"-MQ")){
            //pass
        }else if (args[i]=="-undef"||
        args[i]=="-nostdinc"||
        args[i]=="-nostdinc++"||
        args[i]=="-MD"||
        args[i]=="-MMD"||
        args[i]=="-MG"||
        args[i]=="-MP"){
            //pass
        }else{
            ret=append(ret,args[i])
        }
    }
    return ret
    
}



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

