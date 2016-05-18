package compile
import (
    "os/exec"
    "proto"
)

const (
    CppLocal string = "cpplocal"
    CppRemote string = "cppremote"
)


type RemoteCompileHandler interface {
    PrepareForCompile() bool
    Cmd() *exec.Cmd
    Response() bool
}



func MakeRemoteCompileHandler(cctype string, features map[string]string, dccp *proto.DaemonSideDccp) RemoteCompileHandler{
    switch cctype {
    case "cc":
        fallthrough
    case "gcc":
        fallthrough
    case "g++":
        return NewGCCRemoteCompileHandler(features,dccp)
    }
    return nil
}




