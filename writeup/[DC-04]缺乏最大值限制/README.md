# [DC-04] 缺乏最大值限制

## 漏洞标签

`maxLimit`;

`Overflow`;

## 漏洞描述

缺乏最大值限制(Lack Maximum Limit)

在list里面，首先没有设置int最大值防止溢出。其次，不应该将最大值限制为int的最大值，更恰当的方式是在结构体中增加一个最大长度限制。

## 漏洞分析

文件：`libs/clist/clist.go`

```golang
type CList struct {
	mtx    sync.RWMutex
	wg     *sync.WaitGroup
	waitCh chan struct{}
	head   *CElement // first element
	tail   *CElement // last element
	len    int       // list length
}

// codes ....

func (l *CList) PushBack(v interface{}) *CElement {
	l.mtx.Lock()

	// Construct a new element
	e := &CElement{
		prev:       nil,
		prevWg:     waitGroup1(),
		prevWaitCh: make(chan struct{}),
		next:       nil,
		nextWg:     waitGroup1(),
		nextWaitCh: make(chan struct{}),
		removed:    false,
		Value:      v,
	}

	// Release waiters on FrontWait/BackWait maybe
	if l.len == 0 {
		l.wg.Done()
		close(l.waitCh)
	}

	l.len++

	// Modify the tail
	if l.tail == nil {
		l.head = e
		l.tail = e
	} else {
		e.SetPrev(l.tail) // We must init e first.
		l.tail.SetNext(e) // This will make e accessible.
		l.tail = e        // Update the list.
	}
	l.mtx.Unlock()
	return e
}
```

此处方法 `PushBack` 会通过 `l.len++` 不断增加 `l.len` 的值，而 `l.len` 是一个 `int` 类型。这里有两个问题，一是其增长不受控制，第二是当其不断增加到最大值后会发生溢出。

## 复现或测试步骤

此处使用test脚本测试

### 使用 go test 脚本测试

(脚本位于`DolphinChain-writeup\reproduce\04_max_limit_lack`)

测试方法如下：

```golang
// XuanMao: bug test
// MaxLength = int(^uint(0) >> 1)
// MaxLength = 9223372036854775807
func TestPanicOnMaxLength(t *testing.T) {
    t.Log(MaxLength)
    l := new(CList)
    l.Init()
    for i := 0; i < MaxLength; i++ {
        l.PushBack(100)
    }
}
```

由于没有设置限制，可以调用`MaxLength`次PushBack，会导致CPU占用急剧上升甚至导致`out of memory`

运行并查看结果

```sh
[root@ max_limit_lack]# go test -v -run=TestPanicOnMaxLength
=== RUN   TestPanicOnMaxLength
fatal error: runtime: out of memory

runtime stack:
runtime.throw(0x6b8803, 0x16)
	/usr/local/go/src/runtime/panic.go:608 +0x72
runtime.sysMap(0xc058000000, 0x4000000, 0x8da0b8)
	/usr/local/go/src/runtime/mem_linux.go:156 +0xc7
runtime.(*mheap).sysAlloc(0x8c0180, 0x4000000, 0x0, 0x4d782b)
	/usr/local/go/src/runtime/malloc.go:619 +0x1c7
runtime.(*mheap).grow(0x8c0180, 0x1, 0x0)
	/usr/local/go/src/runtime/mheap.go:920 +0x42
runtime.(*mheap).allocSpanLocked(0x8c0180, 0x1, 0x8da0c8, 0x0)
	/usr/local/go/src/runtime/mheap.go:848 +0x337
runtime.(*mheap).alloc_m(0x8c0180, 0x1, 0x5, 0xc000448000)
	/usr/local/go/src/runtime/mheap.go:692 +0x119
runtime.(*mheap).alloc.func1()
	/usr/local/go/src/runtime/mheap.go:759 +0x4c
runtime.(*mheap).alloc(0x8c0180, 0x1, 0x10005, 0x45ba31)
	/usr/local/go/src/runtime/mheap.go:758 +0x8a
runtime.(*mcentral).grow(0x8c15f8, 0x0)
	/usr/local/go/src/runtime/mcentral.go:232 +0x94
runtime.(*mcentral).cacheSpan(0x8c15f8, 0xc057fff040)
	/usr/local/go/src/runtime/mcentral.go:106 +0x2f8
runtime.(*mcache).refill(0x7f35688d7000, 0xc000047b05)
	/usr/local/go/src/runtime/mcache.go:122 +0x95
runtime.(*mcache).nextFree.func1()
	/usr/local/go/src/runtime/malloc.go:749 +0x32
runtime.systemstack(0x0)
	/usr/local/go/src/runtime/asm_amd64.s:351 +0x66
runtime.mstart()
	/usr/local/go/src/runtime/proc.go:1229

goroutine 6 [running]:
runtime.systemstack_switch()
	/usr/local/go/src/runtime/asm_amd64.s:311 fp=0xc000047c30 sp=0xc000047c28 pc=0x459950
runtime.(*mcache).nextFree(0x7f35688d7000, 0x5, 0x8, 0x65d7e0, 0x897a06)
	/usr/local/go/src/runtime/malloc.go:748 +0xb6 fp=0xc000047c88 sp=0xc000047c30 pc=0x40c036
runtime.mallocgc(0x8, 0x65cf20, 0x6edc00, 0xc057facfc0)
	/usr/local/go/src/runtime/malloc.go:879 +0x5b4 fp=0xc000047d28 sp=0xc000047c88 pc=0x40c7a4
runtime.convT2E64(0x65cf20, 0x18, 0x65d7e0, 0xc057facfc0)
	/usr/local/go/src/runtime/iface.go:324 +0x63 fp=0xc000047d58 sp=0xc000047d28 pc=0x40a283
testing.(*common).decorate(0xc0000d6100, 0xc057febea0, 0x64, 0xc0000b4000, 0xc00aafd300)
	/usr/local/go/src/testing/testing.go:419 +0x1da fp=0xc000047e90 sp=0xc000047d58 pc=0x4d79ba
testing.(*common).log(0xc0000d6100, 0xc057febea0, 0x64)
	/usr/local/go/src/testing/testing.go:597 +0x80 fp=0xc000047f00 sp=0xc000047e90 pc=0x4d89b0
testing.(*common).Log(0xc0000d6100, 0xc000047f68, 0x1, 0x1)
	/usr/local/go/src/testing/testing.go:604 +0x61 fp=0xc000047f38 sp=0xc000047f00 pc=0x4d8b31
github.com/XuanMaoSecLab/DolphinChain/reproduce/max_limit_lack.TestPanicOnMaxLength(0xc0000d6100)
	/home/Go_work/src/github.com/XuanMaoSecLab/DolphinChain/reproduce/max_limit_lack/clist_test.go:24 +0x14f fp=0xc000047fa8 sp=0xc000047f38 pc=0x635f3f
testing.tRunner(0xc0000d6100, 0x6c3658)
	/usr/local/go/src/testing/testing.go:827 +0xbf fp=0xc000047fd0 sp=0xc000047fa8 pc=0x4d980f
runtime.goexit()
	/usr/local/go/src/runtime/asm_amd64.s:1333 +0x1 fp=0xc000047fd8 sp=0xc000047fd0 pc=0x45ba31
created by testing.(*T).Run
	/usr/local/go/src/testing/testing.go:878 +0x353

goroutine 1 [chan receive, 1 minutes]:
testing.(*T).Run(0xc0000d6100, 0x6b7aaa, 0x14, 0x6c3658, 0x477106)
	/usr/local/go/src/testing/testing.go:879 +0x37a
testing.runTests.func1(0xc0000d6000)
	/usr/local/go/src/testing/testing.go:1119 +0x78
testing.tRunner(0xc0000d6000, 0xc000085e08)
	/usr/local/go/src/testing/testing.go:827 +0xbf
testing.runTests(0xc00000c2e0, 0x8b6100, 0x4, 0x4, 0x40c4df)
	/usr/local/go/src/testing/testing.go:1117 +0x2aa
testing.(*M).Run(0xc0000aa100, 0x0)
	/usr/local/go/src/testing/testing.go:1034 +0x165
main.main()
	_testmain.go:48 +0x13d

goroutine 5 [syscall, 1 minutes]:
os/signal.signal_recv(0x0)
	/usr/local/go/src/runtime/sigqueue.go:139 +0x9c
os/signal.loop()
	/usr/local/go/src/os/signal/signal_unix.go:23 +0x22
created by os/signal.init.0
	/usr/local/go/src/os/signal/signal_unix.go:29 +0x41
exit status 2
FAIL	github.com/XuanMaoSecLab/DolphinChain/reproduce/max_limit_lack	77.715s

```

## 修复

修复方法：

```golang
type CList struct {
    mtx    sync.RWMutex
    wg     *sync.WaitGroup
    waitCh chan struct{}
    head   *CElement
    tail   *CElement
    len    int
    maxLen int       // XuanMao: fixed
}

    // XuanMao: fixed 
    // @func (l *CList) PushBack(v interface{}) *CElement
    if l.len >= l.maxLen {
        panic(fmt.Sprintf("clist: maximum length list reached %d", l.maxLen))
    }
```

本漏洞相关修复可以参考 : [Fix](https://github.com/tendermint/tendermint/pull/2289/commits/25b2475b063e155a06ffde0669a01723a83c4808)

## 相关资料

本漏洞设计参考自 : [clist 溢出](https://github.com/tendermint/tendermint/pull/2289/commits/b815efcab3a8bd3f4cca77e65dbe181fa814d8f4)

本漏洞相关 `Issue` 见 : [Issue](https://github.com/tendermint/tendermint/pull/2289)
