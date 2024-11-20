package interfaces

import "time"

type CacheStrategy interface {
	Get(string) (Value, *time.Time, bool)
	Add(string, Value)
	CleanUp(ttl time.Duration)
	Len() int
}

type Value interface {
	Len() int
}
type Entry struct {
	Key      string
	Value    Value
	UpdateAt *time.Time
}

func (ele *Entry) Expired(duration time.Duration) (ok bool) {
	if ele.UpdateAt == nil {
		ok = false
	} else {
		//上次更新时间加上duration在当前时间之前，说明已经过期，ok为true
		ok = ele.UpdateAt.Add(duration).Before(time.Now())
	}
	return ok
}
func (ele *Entry) Touch() {
	//ele.UpdateAt=time.Now()
	nowTime := time.Now()
	ele.UpdateAt = &nowTime
}
