package loadgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"gopcp.v2/chapter4/loadgen/lib"
	"gopcp.v2/helper/log"
)

// 日志记录器。
var logger = log.DLogger()

// myGenerator 代表载荷发生器的实现类型。
type myGenerator struct {
	caller      lib.Caller           // 调用器。
	timeoutNS   time.Duration        // 处理超时时间，单位：纳秒。
	lps         uint32               // 每秒载荷量。
	durationNS  time.Duration        // 负载持续时间，单位：纳秒。
	concurrency uint32               // 载荷并发量。
	tickets     lib.GoTickets        // Goroutine票池。
	ctx         context.Context      // 上下文。
	cancelFunc  context.CancelFunc   // 取消函数。
	callCount   int64                // 调用计数。
	status      uint32               // 状态。
	resultCh    chan *lib.CallResult // 调用结果通道。
}

// NewGenerator 会新建一个载荷发生器。
func NewGenerator(pset ParamSet) (lib.Generator, error) {

	logger.Infoln("New a load generator...")
	if err := pset.Check(); err != nil {
		return nil, err
	}
	gen := &myGenerator{
		caller:     pset.Caller,
		timeoutNS:  pset.TimeoutNS,
		lps:        pset.LPS,
		durationNS: pset.DurationNS,
		status:     lib.STATUS_ORIGINAL,
		resultCh:   pset.ResultCh,
	}
	if err := gen.init(); err != nil {
		return nil, err
	}
	return gen, nil
}

// 初始化载荷发生器。
func (gen *myGenerator) init() error {
	var buf bytes.Buffer
	buf.WriteString("Initializing the load generator...")
	// 载荷的并发量 ≈ 载荷的响应超时时间 / 载荷的发送间隔时间
	var total64 = int64(gen.timeoutNS)/int64(1e9/gen.lps) + 1
	if total64 > math.MaxInt32 {
		total64 = math.MaxInt32
	}
	gen.concurrency = uint32(total64)
	tickets, err := lib.NewGoTickets(gen.concurrency)
	if err != nil {
		return err
	}
	gen.tickets = tickets

	buf.WriteString(fmt.Sprintf("Done. (concurrency=%d)", gen.concurrency))
	logger.Infoln(buf.String())
	return nil
}

// callOne 会向载荷承受方发起一次调用。
func (gen *myGenerator) callOne(rawReq *lib.RawReq) *lib.RawResp {
	atomic.AddInt64(&gen.callCount, 1)
	if rawReq == nil {
		return &lib.RawResp{ID: -1, Err: errors.New("Invalid raw request.")}
	}
	start := time.Now().UnixNano()
	resp, err := gen.caller.Call(rawReq.Req, gen.timeoutNS)
	end := time.Now().UnixNano()
	elapsedTime := time.Duration(end - start)
	var rawResp lib.RawResp
	if err != nil {
		errMsg := fmt.Sprintf("Sync Call Error: %s.", err)
		rawResp = lib.RawResp{
			ID:     rawReq.ID,
			Err:    errors.New(errMsg),
			Elapse: elapsedTime}
	} else {
		rawResp = lib.RawResp{
			ID:     rawReq.ID,
			Resp:   resp,
			Elapse: elapsedTime}
	}
	return &rawResp
}

// asyncSend 会异步地调用承受方接口。
func (gen *myGenerator) asyncCall() {
	gen.tickets.Take() // 对goroutine 票池中的票进行获得和归还操作
	go func() {        // 这个专用的goroutine总数会由goroutine票池控制
		defer func() {
			if p := recover(); p != nil {
				err, ok := interface{}(p).(error) //类型断言表达式判断p的实际类型
				var errMsg string
				if ok {
					errMsg = fmt.Sprintf("Async Call Panic! (error: %s)", err)
				} else {
					errMsg = fmt.Sprintf("Async Call Panic! (clue: %#v)", p)
				}
				logger.Errorln(errMsg)
				result := &lib.CallResult{
					ID:   -1,
					Code: lib.RET_CODE_FATAL_CALL,
					Msg:  errMsg}
				gen.sendResult(result)   
			}
			gen.tickets.Return()
		}()
		rawReq := gen.caller.BuildReq()  // 生成载荷
		// 调用状态：0-未调用或调用中；1-调用完成；2-调用超时。
		var callStatus uint32
		timer := time.AfterFunc(gen.timeoutNS, func() {
			if !atomic.CompareAndSwapUint32(&callStatus, 0, 2) {
				return   // 检查并设置callStatus变量的值，该函数会返回一个bool类型值，用以表示比较并交换是否成功，如果失败，就说明载荷响应接收操作已先完成，忽略超时处理
			}
			result := &lib.CallResult{
				ID:     rawReq.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_WARNING_CALL_TIMEOUT,
				Msg:    fmt.Sprintf("Timeout! (expected: < %v)", gen.timeoutNS),
				Elapse: gen.timeoutNS,
			}
			gen.sendResult(result)
		})
		rawResp := gen.callOne(&rawReq)
		if !atomic.CompareAndSwapUint32(&callStatus, 0, 1) {
			return  // 一旦调用操作的callOne方法返回，就要先检查并设置callStatus变量的值。如果CAS操作不成功，就说明调用操作已超时，忽略后续的响应处理
		}
		timer.Stop()
		var result *lib.CallResult
		if rawResp.Err != nil {
			result = &lib.CallResult{
				ID:     rawResp.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_ERROR_CALL,
				Msg:    rawResp.Err.Error(),
				Elapse: rawResp.Elapse}
		} else {
			result = gen.caller.CheckResp(rawReq, *rawResp)
			result.Elapse = rawResp.Elapse
		}
		gen.sendResult(result)
	}()
}

// sendResult 用于发送调用结果。
func (gen *myGenerator) sendResult(result *lib.CallResult) bool {
	if atomic.LoadUint32(&gen.status) != lib.STATUS_STARTED {
		gen.printIgnoredResult(result, "stopped load generator")
		return false
	}
	select {
	case gen.resultCh <- result:
		return true
	default:
		gen.printIgnoredResult(result, "full result channel")
		return false
	}
}

// printIgnoredResult 打印被忽略的结果。
func (gen *myGenerator) printIgnoredResult(result *lib.CallResult, cause string) {
	resultMsg := fmt.Sprintf(
		"ID=%d, Code=%d, Msg=%s, Elapse=%v",
		result.ID, result.Code, result.Msg, result.Elapse)
	logger.Warnf("Ignored result: %s. (cause: %s)\n", resultMsg, cause)
}

// prepareStop 用于为停止载荷发生器做准备。
func (gen *myGenerator) prepareToStop(ctxError error) {
	logger.Infof("Prepare to stop load generator (cause: %s)...", ctxError)
	atomic.CompareAndSwapUint32(
		&gen.status, lib.STATUS_STARTED, lib.STATUS_STOPPING)  // prepareToStop函数仅在状态为已启动时，把它变为正在停止状态
	logger.Infof("Closing result channel...")
	close(gen.resultCh)
	atomic.StoreUint32(&gen.status, lib.STATUS_STOPPED)  // 变更两次状态的原因是
}

// genLoad 会产生载荷并向承受方发送。
func (gen *myGenerator) genLoad(throttle <-chan time.Time) {
	for {  // 使用一个for循环周期性的向被测试软件发送载荷，这个周期的长短由节流阀控制 
		select {
		case <-gen.ctx.Done():
			gen.prepareToStop(gen.ctx.Err())
			return
		default:
		}
		gen.asyncCall() // 让控制流程和调用流程分离开来，真正实现了载荷发送操作的异步性和并发性
		if gen.lps > 0 {
			select {
			case <-throttle:  //一旦等到节流阀的到期通知，就立即开始下一次迭代
			case <-gen.ctx.Done():   // 接收到上下文的信号，就需要立即为停止载荷发生器做准备
				gen.prepareToStop(gen.ctx.Err())
				return
			}
		}
	}
}

// Start 会启动载荷发生器。
func (gen *myGenerator) Start() bool {
	logger.Infoln("Starting load generator...")
	// 检查是否具备可启动的状态，顺便设置状态为正在启动
	if !atomic.CompareAndSwapUint32(
		&gen.status, lib.STATUS_ORIGINAL, lib.STATUS_STARTING) {
		if !atomic.CompareAndSwapUint32(
			&gen.status, lib.STATUS_STOPPED, lib.STATUS_STARTING) {
			return false
		}
	}

	// 设定节流阀。
	var throttle <-chan time.Time
	if gen.lps > 0 {
		interval := time.Duration(1e9 / gen.lps)
		logger.Infof("Setting throttle (%v)...", interval)
		throttle = time.Tick(interval)   
	}

	// 初始化上下文和取消函数。 目的是让载荷发生器能够在运行一段时间之后自己停下来
	gen.ctx, gen.cancelFunc = context.WithTimeout(
		context.Background(), gen.durationNS)

	// 初始化调用计数。
	gen.callCount = 0

	// 设置状态为已启动。
	atomic.StoreUint32(&gen.status, lib.STATUS_STARTED)

	go func() {
		// 生成并发送载荷。
		logger.Infoln("Generating loads...")
		gen.genLoad(throttle)
		logger.Infof("Stopped. (call count: %d)", gen.callCount)
	}()
	return true
}
//手动的停止载荷发生器
func (gen *myGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(  // 首先检查载荷发生器的状态
		&gen.status, lib.STATUS_STARTED, lib.STATUS_STOPPING) {
		return false
	}
	gen.cancelFunc()
	for {
		if atomic.LoadUint32(&gen.status) == lib.STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return true
}

func (gen *myGenerator) Status() uint32 {
	return atomic.LoadUint32(&gen.status)
}

func (gen *myGenerator) CallCount() int64 {
	return atomic.LoadInt64(&gen.callCount)
}
