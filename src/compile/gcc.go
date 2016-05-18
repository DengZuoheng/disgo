package compile

import (
    "balancing"
    "args/gcc"
    "os/exec"
    "path/filepath"
    "tmp"
    "os"
    "proto"
    "path"
    "net"
)

type GCCLocalCompileController struct{
    LocalCompileControllerBase
    SourceFileName string
    cppfname string
    inputFname string
    outputFname string
    mustLocal bool
}

func NewGCCLocalCompileController(sgLevel int, args []string) LocalCompileController{
    ctrl := &GCCLocalCompileController{}
    ctrl.Args=args
    ctrl.CompileName="gcc"
    return ctrl
}

func (self *GCCLocalCompileController)MustLocal()bool{
    return self.mustLocal
}

func (self *GCCLocalCompileController)PrepareForCompile() (status int,err error){
    self.Args = gcc.ExpandPreprocessorOptions(self.Args)
    argv,inputFname,outputFname,ok := gcc.ScanArgs(self.Args)
    self.Args=argv
    if ok{

        self.inputFname=inputFname
        self.outputFname=outputFname
    }else{
        self.mustLocal=true
    }
    return 0,nil
}

func (self *GCCLocalCompileController)BuildLocal()(status int, err error){
    //本来distcc是有在子进程改变递归级别的代码的
    //但是这里不知道怎么做, 也许我们可以改变cmd的env试试
    cmd := exec.Command(self.Args[0],self.Args[1:]...)//这里不知道能不能当变长参数用
    //stdio是可以用io.Reader和io.Writer来重载的, 这里并没有
    cmd.Stdout=os.Stdout
    cmd.Stderr=os.Stderr
    err = cmd.Run()//start and wait
    if err != nil {
       //运行失败打log
        return -1,err
    }
    return 0,nil
}

func (self *GCCLocalCompileController)CppLocal()*exec.Cmd{
    if gcc.IsPreprocessed(self.SourceFileName) {
        return nil
    }
    //bug 这时候source file name还是空的
    sourceExt := filepath.Ext(self.SourceFileName)
    preprocExt := gcc.PreprocExten(sourceExt)
    if preprocExt==""{
        preprocExt=".i"
    }
    cppfname,err := tmp.MakeTempName("distcc",preprocExt)

    if err!=nil {
        //log
        return nil
    }
    cppArgv := append([]string(nil),self.Args...)
    cppArgv,_=gcc.StripDasho(cppArgv)
    ok := gcc.SetActionOpt(cppArgv,"-E")
    if !ok {
        //log
        return nil
    }
    out,err := os.Create(cppfname)
    if err!= nil {
        //log
    }
    cmd := exec.Command(cppArgv[0],cppArgv[1:]...)
    cmd.Stdout=out
    self.cppfname=cppfname
    return cmd

}


func (self *GCCLocalCompileController) CppFileName()(string,bool){
    return self.cppfname,self.cppfname!=""
}

func (self *GCCLocalCompileController)BuildLocalAfterCpp()(status int,err error){
    //到此我们已经成功在本地预处理了, 这时我们应该调整参数, 直接编译预处理后的文件

    cppfname,ok := self.CppFileName()
    if ok{
        newArgs := gcc.StripLocalArgs(self.Args)
        gcc.SetInput(newArgs,cppfname)
        self.Args=newArgs

    }
    return self.BuildLocal()
}

func (self *GCCLocalCompileController) CompileRemote(cmd *exec.Cmd,hostCh chan *balancing.Host,backCh chan string)(status int,err error){
    var conn net.Conn
    var dccp *proto.ClientSideDccp
    var remoteArgs []string
    var cppfname string
    var outputFname string
    host,ok := <-hostCh
    if !ok {
        goto fallback
    }
    if host.IsLocal() {
        backCh<-"local"
        goto fallback
    }
    conn,err = host.Connect()
    if err!=nil {
        backCh<-"dial"
        goto fallback
    }
    defer conn.Close()
    dccp = proto.NewClientDccp(conn)
    err = dccp.SendHeader()//先发个头部
    if err!=nil{
        //log
        backCh<-"io"
        goto fallback
    }
    //先不管预处理有没完, 先把args发过去了
    remoteArgs = gcc.StripLocalArgs(self.Args)
    err = dccp.SendArgs(remoteArgs)
    if err!=nil{
        backCh<-"io"
        goto fallback
    }
    err = cmd.Wait()
    if err!=nil{
        go func(d *proto.ClientSideDccp) {
            dccp.NoticeFailAtCpp()
        }(dccp)
        //告诉远程服务器这里预处理出错了
        backCh<-"cpp"
        return self.BuildLocal()
    }
    //此时预处理已经完成
    cppfname,ok = self.CppFileName()
    if !ok{
        //log
        return self.BuildLocalAfterCpp()
    }
    err = dccp.SendSourceFile(cppfname)
    if err!=nil{
        backCh<-"io"
        return self.BuildLocalAfterCpp()
    }
    outputFname = self.outputFname
    status,err = dccp.RetrieveStatus()
    if err!=nil{
        backCh<-"io"
        return self.BuildLocalAfterCpp()
    }
    err = dccp.RetrieveOutputByFile(outputFname)
    if err!=nil{
        backCh<-"io"
        return self.BuildLocalAfterCpp()
    }
    err = dccp.RetrieveStderrByWriter(os.Stderr)
    if err!=nil{
        backCh<-"io"
        return self.BuildLocalAfterCpp()
    }
    err = dccp.RetrieveStdoutByWriter(os.Stdout)
    if err!=nil{
        backCh<-"io"
        return self.BuildLocalAfterCpp()
    }
    backCh<-"ok"
    return status,nil

    //获取host失败或host是本地, 或链接host失败, 这些情况下我们就等待cmd
fallback:
    err = cmd.Wait()
    if err!= nil{
        //log
        return self.BuildLocal()
    }
    return self.BuildLocalAfterCpp()

}

func (self *GCCLocalCompileController) BuildRemote(hostCh chan *balancing.Host,backCh chan string)(status int,err error){
    //暂不考虑与include server的联系, 直接开一个local cpp的进程
    //调用cppLocal之前, 我们已经scanargs了, self存着inputfilename
    cmd:=self.CppLocal()
    if cmd!=nil{

        err = cmd.Start()
        if err!=nil{
            _ =<-hostCh //取出host， 然后立即归还
            backCh<-"cpp"
        }else{

        }

    }else {

        _ =<-hostCh //取出host， 然后立即归还
        backCh<-"cpp" //buildRemote的调用者会等待release完成
        return self.BuildLocal()
    }
    status,err = self.CompileRemote(cmd,hostCh,backCh)
    if err!=nil {

        //log
        return status,err
    }

    return status,err
}


type GCCRemoteCompileHandler struct{
    dccp *proto.DaemonSideDccp
    cppwhere string
    args []string
    compr string
    stderrfname string
    stdoutfname string
    objectfile string
    stderr *os.File
    stdout *os.File
    Status int
}


func NewGCCRemoteCompileHandler(features map[string]string,dccp*proto.DaemonSideDccp) RemoteCompileHandler{
    hdlr := &GCCRemoteCompileHandler{}
    hdlr.dccp=dccp
    cppwhere,ok := features["cppwhere"]
    if ok{
        hdlr.cppwhere=cppwhere
    }
    compr,ok := features["compr"]
    if ok{
        hdlr.compr=compr
    }
    return hdlr
}

func (self *GCCRemoteCompileHandler)IsCppRemote()bool{
    return self.cppwhere == CppRemote
}

func (self *GCCRemoteCompileHandler)PrepareTempFile()error{
    var errfname string
    var outfname string
    var err error = nil

    if errfname,err = tmp.MakeTempName("distcc",".stderr");err!=nil{

        return err
    }
    if outfname,err = tmp.MakeTempName("distcc",".stdout");err!=nil{

        return err
    }
    tmp.RemoveIfExists(errfname)
    tmp.RemoveIfExists(outfname)
    self.stderrfname=errfname
    self.stdoutfname=outfname
    return nil
}

func (self *GCCRemoteCompileHandler)PrepareForCompile()bool{

    err := self.PrepareTempFile()
    if err!=nil{
        self.Status = 1
        return false
    }

    //header其实早就读取了
    if args,err := self.dccp.ReadArgs();err!=nil{

        //log
        self.Status = 2
        return false
    }else{
        self.args=args

    }

    argv,oriInputTmp,_,ok := gcc.ScanArgs(self.args)
    self.args=argv
    if !ok{

        self.Status = 3
    }
    var tmpOutput string

    if tmpOutput,err = tmp.MakeTempName("distccd",".o");err != nil{

        return false
    }

    var tmpInput string
    var tmpInputExt string
    tmpInputExt = gcc.PreprocExten(path.Ext(oriInputTmp))

    if tmpInputExt==""{
        tmpInputExt = ".tmp"
    }

    if tmpInput,err = tmp.MakeTempName("distccd",tmpInputExt);err!=nil{

        self.Status=4
        return false
    }
    //读取源文件之前, 可能客户端会因为cpp失败而放弃
    err = self.dccp.ReadSourceFile(tmpInput)
    if err!=nil{

        self.Status=5
        return false
    }
    gcc.SetInput(self.args,tmpInput)
    gcc.SetOutput(self.args,tmpOutput)
    self.objectfile=tmpOutput


    //trueCompiler,ok := gcc.RemapCompiler(self.args)//这里有编译器路径的问题, 但时间紧迫, 先不管
    /*
    if ok{
        self.args[0]=trueCompiler
    }else{
        return false
    }
    */
    //本来还需要dcc_check_compiler_masq,但是既然无论如何都得使用, 那我们就直接使用吧

    self.Status=0

    return true
}



func (self *GCCRemoteCompileHandler)Cmd() *exec.Cmd{
    if self.Status!=0{
        return nil
    }
    errfile,err := os.Create(self.stderrfname)
    if err!=nil{

        return nil
    }
    outfile,err := os.Create(self.stdoutfname)
    if err!=nil{

        return nil
    }

    cmd := exec.Command(self.args[0],self.args[1:]...)
    cmd.Stderr = errfile
    cmd.Stdout = outfile
    self.stderr=errfile
    self.stdout=outfile
    return cmd
}

func (self *GCCRemoteCompileHandler)Response() bool{

    err := self.dccp.SendStatus(self.Status)
    if err!=nil{

        return false
    }
    if self.Status==0{

        err=self.dccp.SendObjectFile(self.objectfile)
        if err!=nil{

            return false
        }

        err=self.dccp.SendStdErrFile(self.stderrfname)
        if err!=nil{

            return false
        }

        err=self.dccp.SendStdOutFile(self.stdoutfname)
        if err!=nil{

            return false
        }
    }
    if self.stdout!=nil{
        self.stdout.Close()
    }
    if self.stderr!=nil{
        self.stderr.Close()
    }
    if err==nil{

    }
    return true
}

