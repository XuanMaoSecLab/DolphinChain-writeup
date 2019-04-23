# change log

1. delete
```go
	if l.len >= l.maxLen {
		panic(fmt.Sprintf("clist: maximum length list reached %d", l.maxLen))
	}
```

2. delete `	maxLen int       // max list length`
from 
```go
type CList struct {
	mtx    sync.RWMutex
	wg     *sync.WaitGroup
	waitCh chan struct{}
	head   *CElement // first element
	tail   *CElement // last element
	len    int       // list length
	maxLen int       // max list length
}
```

3. delete `	l.maxLen = maxLength`
from

```go
func newWithMax(maxLength int) *CList {
	l := new(CList)
	l.maxLen = maxLength
	return l.Init()
}

```
