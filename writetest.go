package main
import (
	"sync"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/dlclark/regexp2"
	_ "github.com/lib/pq"
	"unsafe"
	"github.com/tidwall/gjson"
	"strconv"
	"time"
	"database/sql"
	"strings"
)

/*
一个标准的协程+信道实现

*/

func main() {
	fmt.Println(time.Now().String())
	taskChan := make(chan int)
	TCount := 100
	var wg sync.WaitGroup //创建一个sync.WaitGroup

	// 产生任务
	go func() {
		for i := 0; i < 9000000; i++ {
			taskChan <- i
		}
		// 全部任务都输入后关闭信道，告诉工作者进程没有新任务了。
		close(taskChan)
	}()

	// 启动 TCount 个协程执行任务
	wg.Add(TCount)
	for i := 0; i < TCount; i++ {

		// 注意：如果协程内使用了 i，必须有这一步，或者选择通过参数传递进协程。
		// 否则 i 会被 for 所在的协程修改，协程实际使用时值并不确定。
		//i := i

		go func() {

			// 协程结束时报告当前协程执行完毕。
			defer func() { wg.Done() }()

			//fmt.Printf("工作者 %v 启动...\r\n", i)

			for task := range taskChan {

				// 建立匿名函数执行任务的目的是为了捕获单个任务崩溃，防止造成整个工作者、系统崩溃。
				func() {

					//defer func() {
						//err := recover()
						//if err != nil {
						//	fmt.Printf("任务失败：工作者i=%v, task=%v, err=%v\r\n", i, task, err)
					//	}
					//}()

					// 故意崩溃，看看是不是会造成整个系统崩溃。
					//if task%100==0{
					//	panic("故意崩溃啦")
					//}

					// 这里的 task 并不需要通过参数传递进来。
					// 原因是这里是同步执行的，并不会被其它协程修改。
					Title(strconv.Itoa(task))
					//fmt.Printf("task=%v\r\n",task)
				}()
			}

			//fmt.Printf("工作者 %v 结束。\r\n", i)
		}()
	}

	//等待所有任务完成
	wg.Wait()
	print("全部任务结束")
	fmt.Println(time.Now().String())
}

func Title(id string){
	req:=fasthttp.AcquireRequest()
	req.SetRequestURI("http://www.bilibili.com/video/av"+id+"/")
	req.Header.SetMethod("GET")
	req.Header.Set("Host","www.bilibili.com")
	req.Header.Set("Content-Length","940")
	req.Header.Set("Accept-Encoding","deflate, sdch")
	req.Header.Set("Cookie", "DedeUserID=2003393; DedeUserID__ckMd5=bf8db768dd87f783; SESSDATA=c169060d%2C1490365767%2C6b4beba9; bili_jct=e04f5fe87366b5a1fce597c72bac5b5a; buvid3=7BA59344-C07D-45DF-870D-BC1FFAFA03719242infoc; pgv_pvi=8812679168; fts=1487773759; _cnt_dyn=null; uTZ=-480; LIVE_BUVID=260bcc987239968ef350cce06a0a855f; LIVE_BUVID__ckMd5=1acc4e3019efb10b; sid=ci270bip; rpdid=iwxllilxwmdopqqxlmwiw; UM_distinctid=15aa2eca37d5d-09597e0cefe322-5e4f2b18-1fa400-15aa2eca37e13; LIVE_LOGIN_DATA=a09b1d4fe41753717a8124d2a2686f6ed590297c; LIVE_LOGIN_DATA__ckMd5=0af009c61df3cf3d; purl_token=bilibili_1489034264; _cnt_pm=0; _cnt_notify=9; DedeID=5462081; user_face=http%3A%2F%2Fi0.hdslb.com%2Fbfs%2Fface%2F8926f0041b151e7961d42f94126b0c1251304ade.jpg; CNZZDATA2724999=cnzz_eid%3D1999797100-1487771208-https%253A%252F%252Fpassport.bilibili.com%252F%26ntime%3D1489044575; _dfcaptcha=b63fc78d66e2879c9d0331b09d8573f8")
	resp:=fasthttp.AcquireResponse()
	fasthttp.Do(req,resp)
	db, _ := sql.Open("postgres", "user=bilibili password=-ggggw159874 dbname=bilibili host=bilibili.chtljhcafylp.ap-northeast-1.rds.amazonaws.com sslmode=disable")
	if len(resp.Body())>7000{
		body,_:=resp.BodyGunzip()
		author := regexp2.MustCompile(`name="author" content=".*?"`, 0)
		title :=regexp2.MustCompile(`<h1 title=".*?"`, 0)
		m, _ := author.FindStringMatch(*(*string)(unsafe.Pointer(&body)))
		 t, _ := title.FindStringMatch(*(*string)(unsafe.Pointer(&body)))
		req.SetRequestURI("http://api.bilibili.com/archive_stat/stat?&aid="+id+"&type=jsonp")
		fasthttp.Do(req,resp)
		value:=gjson.GetBytes(resp.Body(),"data").Map()
		stmt, _ := db.Prepare("INSERT INTO video(name,author,view,danmuku,reply,favourite,coin,share,avid) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)")
		stmt.Exec(strings.Replace(t.String(),`<h1 title=`,"",-1),strings.Replace(m.String(),`name="author" content=`,"",-1),value["view"].String(),value["danmaku"].String(),value["reply"].String(),value["favorite"].String(),value["coin"].String(),value["share"].String(),id)
	}
	defer db.Close()
}