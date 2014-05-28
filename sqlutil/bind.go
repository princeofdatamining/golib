
package sqlutil

import (
    "reflect"
)

type bindObj struct {
    query       string
    argFields   []string
    keyFields   []string
    setFields   []string
    argValues   []interface{}
}
func (this *bindObj) bindArgs(obj reflect.Value) (err error) {
    if this.argValues == nil {
        this.argValues = make([]interface{}, len(this.argFields))
    }
    for i, fn := range this.argFields {
        this.argValues[i] = obj.FieldByName(fn).Interface()
    }
    return 
}

func triggerArg(exec SQLExecutor) ([]reflect.Value) {
    return []reflect.Value{ reflect.ValueOf(exec) }
}

func triggerRun(method string, sender reflect.Value, arg []reflect.Value) (err error) {
    m := sender.MethodByName(method)
    if !m.IsValid() {
        return 
    }
    ret := m.Call(arg)
    if len(ret) <= 0 || ret[0].IsNil() {
        return 
    }
    return ret[0].Interface().(error)
}
