package balancing

import (
    "net"
    "math/rand"
    "container/list"
    "sort"
    "fmt"
)


type DB struct{
    AllHost map[string]*Host
    AvailHost map[string]*Host
    UnavailHost map[string]*Host
    SilentHost map[string]*Host
    OverloadedHost map[string]*Host
    AvailSlots map[int]*list.List
    AssignedSlots map[string]string
    numAssignedSlots int
    numConcurrentAssigned int
}

var dbInstance *DB = nil

func newDB()*DB{
    db := &DB{}
    db.AllHost = make(map[string]*Host,0) //map["host:port"]*Host
    db.AvailHost=make(map[string]*Host,0)
    db.UnavailHost=make(map[string]*Host,0)
    db.SilentHost = make(map[string]*Host,0)
    db.OverloadedHost = make(map[string]*Host,0)
    db.AvailSlots = make(map[int]*list.List,0)//map[tier]=list<"host:port">
    db.AssignedSlots = make(map[string]string,0)//map[conn.Addr().String()]="host:port"
    db.numAssignedSlots=0
    db.numConcurrentAssigned=0
    return db
}

func DBInstance() *DB{
    if dbInstance != nil{
        return dbInstance
    }
    dbInstance = newDB()
    return dbInstance
}

func (self *DB)GetHost(addr string)(host *Host,ok bool){
    host,ok = self.AllHost[addr]
    return
}

func (self *DB)HaveHost(addr string)bool{
    _,ok := self.AllHost[addr]
    return ok
}

func (self *DB)GetBestAvailSlot()(string,bool){
    //遍历availcpus, 找到第一个不为空的, 在里面随机挑一个, 然后移除之
    var slot string
    var keys []int
    for k:=range self.AvailSlots{
        keys = append(keys,k)
    }
    sort.Sort(sort.Reverse(sort.IntSlice(keys)))
    for _,k:=range keys{

        slotList := self.AvailSlots[k]
        nSlot := slotList.Len()
        fmt.Printf("in tier %d we have %d slots\n",k,nSlot)
        if nSlot == 0{
            continue
        }
        seletced := rand.Intn(nSlot)
        fmt.Println("rand select:",seletced)
        e := slotList.Front()
        for i:=0;i<seletced;i++{
            e = e.Next()
        }
        slot,_ = e.Value.(string)
        slotList.Remove(e)
        fmt.Println("we got slot in tier:",k)
        return slot,true
    }
    return slot,false
}

func (self *DB)AssignCpuToClient(slot string,client net.Conn){
    str := client.RemoteAddr().String()
    self.AssignedSlots[str] = slot
    self.numAssignedSlots++
    if t:= len(self.AssignedSlots);t>self.numConcurrentAssigned{
        self.numConcurrentAssigned = t
    }

}

func (self *DB)ReleaseSlot(client net.Conn)bool{
    str := client.RemoteAddr().String()
    slot,ok := self.AssignedSlots[str]
    if ok{
        host,ok := self.GetHost(slot)
        if !ok{
            return false
        }
        if host.IsAvail(){
            tier := host.GetTier()
            self.AddSlotToTier(tier,slot,1)
            return true
        }
    }
    return false
}

func (self *DB)AddNewHost(host *Host){
    self.AddToHostSet(self.AllHost,host)
    self.AddToAvailDb(host)
}

func (self *DB)AddToHostSet(set map[string]*Host, host *Host)bool{
    _,ok := set[host.RawAddr]
    if ok{
        //log
    }
    set[host.RawAddr]=host
    //log
    return ok
}


func (self *DB)DelFromHostSet(set map[string]*Host, host *Host)bool{
    delete(set,host.RawAddr)
    return true
    //log
}

func (self *DB)AddToAvailDb(host *Host){
    self.AddToHostSet(self.AvailHost,host)
    self.AddSlotToTier(host.GetTier(),host.RawAddr,host.GetNumSlots())

}

func (self *DB)DelFromAvailDb(host *Host) bool{
    slotList,ok := self.AvailSlots[host.GetTier()]
    if ok {
        var next *list.Element
        //移除所有相同host的slot
        for e:=slotList.Front();e!=nil;e=next{
            next = e.Next()
            val := e.Value.(string)
            if val==host.RawAddr{
                slotList.Remove(e)
            }
        }
    }
    return self.DelFromHostSet(self.AvailHost,host)
}

func (self *DB)AddToOverloadedDb(host *Host){
    self.AddToHostSet(self.OverloadedHost,host)
}

func (self *DB)DelFromOverloadedDb(host *Host){
    self.DelFromHostSet(self.OverloadedHost,host)
}

func (self *DB)AddToSilentDb(host *Host){
    self.AddToHostSet(self.SilentHost,host)
}

func (self *DB)DelFromSilentDb(host *Host){
    self.DelFromHostSet(self.SilentHost,host)
}

func (self *DB)AddToUnavailDb(host *Host){
    self.AddToHostSet(self.UnavailHost,host)
}

func (self *DB)DelFromUnavailDb(host *Host){
    self.DelFromHostSet(self.UnavailHost,host)
}

func (self *DB)AddSlotToTier(tier int, slot string, numCpus int){
    slotList, ok := self.AvailSlots[tier]
    if !ok{
        slotList = list.New()
        self.AvailSlots[tier] = slotList
    }
    for i:=0;i<numCpus;i++{
           slotList.PushBack(slot)
    }
}



func (self *DB)DelSlotFromTier(tier int, slot string)int{
    slotList,ok := self.AvailSlots[tier]
    count := 0
    if ok {
        var next *list.Element
        for e:= slotList.Front();e!=nil;e=next{
            next=e.Next()
            val := e.Value.(string)
            if val == slot{
                slotList.Remove(e)
                count++
            }
        }
    }
    return count
}

func (self *DB)MoveCpus(host *Host,oldTier,newTier int){
    numCpusDel := self.DelSlotFromTier(oldTier,host.RawAddr)
    self.AddSlotToTier(newTier,host.RawAddr,numCpusDel)
}

func (self *DB)HandleSilentHosts(){
    for _,host := range self.AllHost {
        if host.SeemsDown(){
            host.Silent()
        }
    }
}
