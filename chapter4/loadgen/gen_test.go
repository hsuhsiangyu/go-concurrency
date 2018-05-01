package loadgen

import (
	"testing"
	"time"

	loadgenlib "gopcp.v2/chapter4/loadgen/lib"
	helper "gopcp.v2/chapter4/loadgen/testhelper"
)

// printDetail 代表是否打印详细结果。
var printDetail = false

func TestStart(t *testing.T) {

	// 初始化服务器。
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8080"
	t.Logf("Startup TCP server(%s)...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server startup failing! (addr=%s)!\n", serverAddr)
		t.FailNow()
	}

	// 初始化载荷发生器。
	pset := ParamSet{
		Caller:     helper.NewTCPComm(serverAddr),  //创建和初始化一个基于TCP协议的调用器
		TimeoutNS:  50 * time.Millisecond,
		LPS:        uint32(1000),
		DurationNS: 10 * time.Second,
		ResultCh:   make(chan *loadgenlib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		pset.TimeoutNS, pset.LPS, pset.DurationNS)
	gen, err := NewGenerator(pset)
	if err != nil {
		t.Fatalf("Load generator initialization failing: %s\n",
			err)
		t.FailNow()
	}

	// 开始！
	t.Log("Start load generator...")
	gen.Start()

	// 显示结果。
	countMap := make(map[loadgenlib.RetCode]int)
	for r := range pset.ResultCh {  // 不断尝试从调用结果通道接收结果
		countMap[r.Code] = countMap[r.Code] + 1  // 按照响应代码分类对响应计数
		if printDetail {
			t.Logf("Result: ID=%d, Code=%d, Msg=%s, Elapse=%v.\n",
				r.ID, r.Code, r.Msg, r.Elapse)
		}
	}

	var total int
	t.Log("RetCode Count:")
	for k, v := range countMap {
		codePlain := loadgenlib.GetRetCodePlain(k)
		t.Logf("  Code plain: %s (%d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}

	t.Logf("Total: %d.\n", total)
	successCount := countMap[loadgenlib.RET_CODE_SUCCESS]
	tps := float64(successCount) / float64(pset.DurationNS/1e9)
	t.Logf("Loads per second: %d; Treatments per second: %f.\n", pset.LPS, tps)
}
// 手动停止
func TestStop(t *testing.T) {

	// 初始化服务器。
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8081"
	t.Logf("Startup TCP server(%s)...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server startup failing! (addr=%s)!\n", serverAddr)
		t.FailNow()
	}

	// 初始化载荷发生器。
	pset := ParamSet{
		Caller:     helper.NewTCPComm(serverAddr),
		TimeoutNS:  50 * time.Millisecond,
		LPS:        uint32(1000),
		DurationNS: 10 * time.Second,
		ResultCh:   make(chan *loadgenlib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		pset.TimeoutNS, pset.LPS, pset.DurationNS)
	gen, err := NewGenerator(pset)
	if err != nil {
		t.Fatalf("Load generator initialization failing: %s.\n",
			err)
		t.FailNow()
	}

	// 开始！
	t.Log("Start load generator...")
	gen.Start()
	timeoutNS := 2 * time.Second
	time.AfterFunc(timeoutNS, func() {
		gen.Stop()      // 在这里设定了手动停止时间 2s 
	})

	// 显示调用结果。
	countMap := make(map[loadgenlib.RetCode]int)
	count := 0
	for r := range pset.ResultCh {
		countMap[r.Code] = countMap[r.Code] + 1
		if printDetail {
			t.Logf("Result: ID=%d, Code=%d, Msg=%s, Elapse=%v.\n",
				r.ID, r.Code, r.Msg, r.Elapse)
		}
		count++
	}

	var total int
	t.Log("RetCode Count:")
	for k, v := range countMap {
		codePlain := loadgenlib.GetRetCodePlain(k)
		t.Logf("  Code plain: %s (%d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}

	t.Logf("Total: %d.\n", total)
	successCount := countMap[loadgenlib.RET_CODE_SUCCESS]
	tps := float64(successCount) / float64(timeoutNS/1e9)
	t.Logf("Loads per second: %d; Treatments per second: %f.\n", pset.LPS, tps)
}
