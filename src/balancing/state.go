package balancing

import "fmt"

type StateCode int

const (
    Unknown StateCode= iota
    Available
    Unavailable
    Overloaded
    Silent
)

type State interface {
    AsCode()StateCode
    Avail(host *Host)
    Unavail(host *Host)
    Silent(host *Host)
    Overloaded(host *Host)
    RemoveFromDB(host *Host)
    AddToDB(host *Host)
}

type StateBase struct{
    state StateCode
}

func (self *StateBase)AsCode()StateCode{
    return self.state
}

func (self *StateBase)Avail(host *Host){

}
func (self *StateBase)Unavail(host *Host){

}
func (self *StateBase)Silent(host *Host){

}
func (self *StateBase)Overloaded(host *Host){

}

func (self *StateBase) RemoveFromDB(host *Host){

}

func (self *StateBase) AddToDB(host *Host){

}

type AvailState struct{
    StateBase
}

type UnavailState struct{
    StateBase
}

type SilentState struct{
    StateBase
}

type OverloadedState struct{
    StateBase
}

var OverloadedInstance *OverloadedState = nil

var SilentInstance *SilentState = nil

var UnavailInstance *UnavailState = nil

var AvailInstance *AvailState = nil

func StateAvailInstance() State{
    if AvailInstance != nil{
        return AvailInstance
    }
    AvailInstance = new(AvailState)
    AvailInstance.state=Available
    return AvailInstance
}

func StateUnavailInstance() State{
    if UnavailInstance != nil{
        return UnavailInstance
    }
    UnavailInstance = new(UnavailState)
    UnavailInstance.state=Unavailable
    return AvailInstance
}

func StateSilentInstance() State{
    if SilentInstance != nil{
        return SilentInstance
    }
    SilentInstance = new(SilentState)
    SilentInstance.state=Silent
    return AvailInstance
}

func StateOverloadedInstance() State{
    if OverloadedInstance != nil{
        return OverloadedInstance
    }
    OverloadedInstance = new(OverloadedState)
    OverloadedInstance.state=Overloaded
    return OverloadedInstance
}

func (self *AvailState) RemoveFromDB(host *Host){
    db := DBInstance()
    db.DelFromAvailDb(host)
}

func (self *AvailState) Unavail(host *Host){
    db := DBInstance()
    db.DelSlotFromTier(host.GetTier(),host.RawAddr)
    self.RemoveFromDB(host)
    host.ChangeState(StateUnavailInstance())

}

func (self *AvailState) Silent(host *Host){
    db := DBInstance()
    db.DelSlotFromTier(host.GetTier(),host.RawAddr)
    self.RemoveFromDB(host)
    host.ChangeState(StateSilentInstance())
}

func (self *AvailState) Overloaded(host *Host){
    db := DBInstance()
    db.DelSlotFromTier(host.GetTier(),host.RawAddr)
    self.RemoveFromDB(host)
    host.ChangeState(StateOverloadedInstance())
}



func (self *AvailState) AddToDB(host *Host){
    fmt.Println("!!!!!!!!!!!!!!!!!!enter avail add to db")
    db := DBInstance()
    db.AddToAvailDb(host)
}

func (self *UnavailState) Avail(host *Host){
    self.RemoveFromDB(host)
    tier := host.GetTier()
    if tier == 0{
        host.ChangeState(StateOverloadedInstance())
    }else{
        host.ChangeState(StateAvailInstance())
    }
}

func (self *UnavailState) AddToDB(host *Host){
    db := DBInstance()
    db.AddToUnavailDb(host)
}

func (self *UnavailState) RemoveFromDB(host *Host){
    db := DBInstance()
    db.DelFromUnavailDb(host)
}

func (self *SilentState) Avail(host *Host){
    self.RemoveFromDb(host)
    host.ChangeState(StateAvailInstance())
}

func (self *SilentState) Unavail(host *Host) {
    self.RemoveFromDb(host)
    host.ChangeState(StateUnavailInstance())
}

func (self *SilentState) AddToDB(host *Host){
    db := DBInstance()
    db.AddToSilentDb(host)
}

func (self *SilentState) RemoveFromDb(host *Host){
    db := DBInstance()
    db.DelFromSilentDb(host)
}

func (self *OverloadedState)Avail(host *Host){
    fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!enter overloaded state avail")
    self.RemoveFromDB(host)
    host.ChangeState(StateAvailInstance())
}

func (self *OverloadedState)Unavail(host *Host){
    self.RemoveFromDB(host)
    host.ChangeState(StateUnavailInstance())
}

func (self *OverloadedState)Silent(host *Host){
    self.RemoveFromDB(host)
    host.ChangeState(StateSilentInstance())
}

func (self *OverloadedState)RemoveFromDB(host *Host){
    db := DBInstance()
    db.DelFromOverloadedDb(host)
}

func (self *OverloadedState)AddToDB(host *Host){
    fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!enter overloaded state add to db")
    db := DBInstance()
    db.AddToOverloadedDb(host)
}