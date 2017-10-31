package pidmonitor

import (
	"bytes"
	"fmt"
	"sync"
	"time"
	"io"
    "os/exec"
	"strings"
	"strconv"
	"net/http"
	"net/url"
	"io/ioutil"
	"os"
)

/*
func main(){
	fmt.Println("start")
	monitor := pidmonitor.New()
	//monitor.WatchPid("top1",16005)
	//monitor.WatchPid("top2",16006)
	//monitor.WatchDir("./pids/")
	//monitor.Run()

	//写入pid
	//monitor.WritePid("./pids/","switch")

	fmt.Println("end")
}
*/

func New(token string) *monitor {
	monitor := new(monitor)
	monitor.token = token
	monitor.pids = make(map[string]int64)
	return monitor
}

type monitor struct{
	token string
	pids map[string]int64
	wg sync.WaitGroup
}

func (m *monitor)Run(){
	fmt.Printf("%v\n",m.pids)
	for name,pid := range m.pids {
		m.wg.Add(1)
		go func(name string,pid int64) {
			defer m.wg.Done()
			tc:=time.Tick(time.Second)
			for{
			    <-tc
			    if !m.pidExist(pid){
			    	text := fmt.Sprintf("%s %s_%d\n",time.Now().Format("2006-01-02 15:04:05"),name,pid)
			    	//fmt.Println(text)
			    	m.notify(text)
			    	break
			    } else {
			    	//fmt.Println("exist")
			    }
			}
		}(name,pid)	
	}
	m.wg.Wait()
}

func (m *monitor)WatchDir(dirPath string) {
	//读取pid
	pids,err := m.readPids(dirPath)
	if err!=nil{
		return
	}

    //监控pid
	for name, pid := range pids {
	    m.pids[name] = pid
    }
    
}

func (m *monitor)WatchPid(name string,pid int64){
	m.pids[name] = pid
}

func (m *monitor)WritePid(dirPath string, fileName string) error {
    err := os.MkdirAll(dirPath, 0777)
    if err != nil {
        return err
    }

    file, err := os.OpenFile(dirPath+fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777) 
    if err != nil {
        return err
    } 
    defer file.Close() 

    _, err = file.WriteString(fmt.Sprintf("%d", os.Getpid())) 
    if err != nil { 
        return err
    } 
    return nil
}

func (m *monitor)readPid(filePath string) (int64, error) { 
    file, err := os.OpenFile(filePath, os.O_RDWR, os.ModeType) 
    if err != nil { 
        return 0,err
    } 
    defer file.Close() 

    b, err := ioutil.ReadAll(file) 
    if err != nil { 
        return 0,err
    } 
    pid, err := strconv.ParseInt(string(b),10,16)
    if err != nil { 
        return 0,err
    }
    return pid,nil
}

func (m *monitor)readPids(dirPath string) (map[string]int64,error) {
	var pids map[string]int64
	pids = make(map[string]int64)

    dirList, err := ioutil.ReadDir(dirPath)
    if err != nil {
        return pids,err
    }

    for _, v := range dirList {
    	pid,_ := m.readPid(dirPath+v.Name())
    	pids[v.Name()]=pid
    }

    return pids,nil
}

func (m *monitor)pidExist(pid int64) bool {
    c1 := exec.Command("ps","-p", fmt.Sprintf("%d", pid))
    c2 := exec.Command("wc", "-l")

    r, w := io.Pipe() 
    c1.Stdout = w
    c2.Stdin = r

    var out bytes.Buffer
    c2.Stdout = &out

    c1.Start()
    c2.Start()
    c1.Wait()
    w.Close()
    c2.Wait()

    num,_ := strconv.ParseInt(strings.TrimSpace(out.String()),10,8)

 	return num>1
}

func (m *monitor)notify(text string) error{
	url := fmt.Sprintf("https://sc.ftqq.com/%s.send?text=%s", url.QueryEscape(m.token), url.QueryEscape(text))
	resp, err := http.Get(url)
	defer resp.Body.Close()
	return err
}