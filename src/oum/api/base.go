package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	js "github.com/bitly/go-simplejson"
)

type API interface {
	Run() *Result
	do() *js.Json
}

type base struct {
	rep  http.ResponseWriter
	req  *http.Request
	errs []error
}

func New(name string, w http.ResponseWriter, r *http.Request) (a API) {
	r.ParseForm()
	b := base{
		rep: w,
		req: r,
	}

	creator := apis[name]
	if creator == nil {
		creator = apis[API_UNDEFINED]
	}

	return creator(name, b)
}

func (b *base) exit(code uint64, msg string, v ...interface{}) *js.Json {
	e := &apiError{
		code: code,
		msg:  fmt.Sprintf(msg, v...),
	}
	panic(e)
	return nil
}

func (b *base) run(do func() *js.Json) (ret *Result) {
	defer func() {
		err := recover()
		if err == nil {
			return
		}

		ret = new(Result)
		switch err.(type) {
		case apiError:
			e := err.(apiError)
			ret.Code = e.code
			ret.Message = e.msg
		case *apiError:
			e := err.(*apiError)
			ret.Code = e.code
			ret.Message = e.msg
		case error:
			e := err.(error)
			ret.Code = ERR_UNEXPECTED
			ret.Message = e.Error()
		case string:
			e := err.(string)
			ret.Code = ERR_UNEXPECTED
			ret.Message = e
		default:
			ret.Code = ERR_UNEXPECTED
			ret.Message = fmt.Sprintf("%#v", err)
		}
	}()

	if do == nil {
		b.exit(ERR_API_INVALID, "Invalid api implementation")
	}

	ret = &Result{
		Data: do(),
	}

	return
}

var pattern_int *regexp.Regexp = regexp.MustCompile(`^[+-]?\d+$`)
var pattern_uint *regexp.Regexp = regexp.MustCompile(`^\d+$`)
var pattern_float *regexp.Regexp = regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`)

func (b *base) oStr(key string, opt ...string) string {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return ""
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return ""
	}
	return ret
}

func (b *base) xStr(key string) string {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	return ret
}

func (b *base) oInt(key string, opt ...int64) int64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return 0
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return 0
	}
	if pattern_int.MatchString(ret) == false {
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be integer", key)
	}
	i, err := strconv.ParseInt(ret, 10, 64)
	if err != nil {
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be integer", key)
	}
	return i
}

func (b *base) xInt(key string) int64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	if pattern_int.MatchString(ret) == false {
		b.exit(ERR_ARG_INVALID, "required arg[%s], invalid format, should be integer", key)
	}
	i, err := strconv.ParseInt(ret, 10, 64)
	if err != nil {
		b.exit(ERR_ARG_INVALID, "required arg[%s], invalid format, should be integer", key)
	}
	return i
}

func (b *base) oUint(key string, opt ...uint64) uint64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return 0
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return 0
	}
	if pattern_uint.MatchString(ret) == false {
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be unsigned integer", key)
	}
	i, err := strconv.ParseUint(ret, 10, 64)
	if err != nil {
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be unsigned integer", key)
	}
	return i
}

func (b *base) xUint(key string) uint64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	if pattern_uint.MatchString(ret) == false {
		b.exit(ERR_ARG_INVALID, "required arg[%s], invalid format, should be unsigned integer", key)
	}
	i, err := strconv.ParseUint(ret, 10, 64)
	if err != nil {
		b.exit(ERR_ARG_INVALID, "required arg[%s], invalid format, should be unsigned integer", key)
	}
	return i
}

func (b *base) oFloat(key string, opt ...float64) float64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return 0
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return 0
	}
	if pattern_float.MatchString(ret) == false {
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be float", key)
	}
	f, err := strconv.ParseFloat(ret, 64)
	if err != nil {
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be float", key)
	}
	return f
}

func (b *base) xFloat(key string) float64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	if pattern_float.MatchString(ret) == false {
		b.exit(ERR_ARG_INVALID, "required arg[%s], invalid format, should be float", key)
	}
	f, err := strconv.ParseFloat(ret, 64)
	if err != nil {
		b.exit(ERR_ARG_INVALID, "required arg[%s], invalid format, should be float", key)
	}
	return f
}

func (b *base) oBool(key string, opt ...bool) bool {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return false
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return false
	}
	ret = strings.ToLower(ret)
	switch ret {
	case "0", "f", "false", "n", "no":
		return false
	case "1", "t", "true", "y", "yes":
		return true
	default:
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be bool", key)
	}
	return false
}

func (b *base) xBool(key string) bool {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	ret := strings.TrimSpace(value[0])
	if len(ret) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret = strings.ToLower(ret)
	switch ret {
	case "0", "f", "false", "n", "no":
		return false
	case "1", "t", "true", "y", "yes":
		return true
	default:
		b.exit(ERR_ARG_INVALID, "optional arg[%s], invalid format, should be bool", key)
	}
	return false
}

func (b *base) oJson(key string, opt ...*js.Json) *js.Json {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return js.New()
	}
	tmp := strings.TrimSpace(value[0])
	if len(tmp) == 0 {
		if len(opt) > 0 {
			return opt[0]
		}
		return js.New()
	}
	ret, err := js.NewJson([]byte(tmp))
	if err != nil {
		b.exit(ERR_ARG_INVALID,
			"optional arg[%s], invalid format, should be json", key)
	}
	return ret
}

func (b *base) xJson(key string) *js.Json {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	tmp := strings.TrimSpace(value[0])
	if len(tmp) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret, err := js.NewJson([]byte(tmp))
	if err != nil {
		b.exit(ERR_ARG_INVALID,
			"required arg[%s], invalid format, should be json", key)
	}
	return ret
}

func (b *base) _set(str string) []string {
	tmp := make(map[string]struct{})
	for _, v := range strings.Split(str, ",") {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			tmp[v] = struct{}{}
		}
	}
	ret := make([]string, 0, len(tmp))
	for k, _ := range tmp {
		ret = append(ret, k)
	}
	return ret
}

func (b *base) _array(str string) []string {
	ret := make([]string, 0, 2)
	for _, v := range strings.Split(str, ",") {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			ret = append(ret, v)
		}
	}
	return ret
}

// actually, string set
func (b *base) oCSV(key string, opts ...string) []string {
	return b.oStrSet(key, opts...)
}

// actually, string set
func (b *base) xCSV(key string) []string {
	return b.xStrSet(key)
}

// string set, do not support "" element
func (b *base) oStrSet(key string, opts ...string) []string {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []string{}
	}
	ret := b._set(value[0])
	if len(ret) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []string{}
	}
	return ret
}

func (b *base) xStrSet(key string) []string {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	ret := b._set(value[0])
	if len(ret) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	return ret
}

func (b *base) oIntSet(key string, opts ...int64) []int64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []int64{}
	}
	tmp := make(map[int64]struct{})
	for _, v := range b._set(value[0]) {
		if pattern_int.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be integer set", key)
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be integer set", key)
		}
		tmp[i] = struct{}{}
	}
	if len(tmp) == 0 {
		if len(opts) > 0 {
			return opts
		}
	}
	ret := make([]int64, 0, len(tmp))
	for k, _ := range tmp {
		ret = append(ret, k)
	}
	return ret
}

func (b *base) xIntSet(key string) []int64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	tmp := make(map[int64]struct{})
	for _, v := range b._set(value[0]) {
		if pattern_int.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be integer set", key)
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be integer set", key)
		}
		tmp[i] = struct{}{}
	}
	if len(tmp) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret := make([]int64, 0, len(tmp))
	for k, _ := range tmp {
		ret = append(ret, k)
	}
	return ret
}

func (b *base) oUintSet(key string, opts ...uint64) []uint64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []uint64{}
	}
	tmp := make(map[uint64]struct{})
	for _, v := range b._set(value[0]) {
		if pattern_uint.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be unsigned integer set", key)
		}
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be unsigned integer set", key)
		}
		tmp[i] = struct{}{}
	}
	if len(tmp) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []uint64{}
	}
	ret := make([]uint64, 0, len(tmp))
	for k, _ := range tmp {
		ret = append(ret, k)
	}
	return ret
}

func (b *base) xUintSet(key string) []uint64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	tmp := make(map[uint64]struct{})
	for _, v := range b._set(value[0]) {
		if pattern_uint.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be unsigned integer set", key)
		}
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be unsigned integer set", key)
		}
		tmp[i] = struct{}{}
	}
	if len(tmp) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret := make([]uint64, 0, len(tmp))
	for k, _ := range tmp {
		ret = append(ret, k)
	}
	return ret
}

func (b *base) oFloatSet(key string, opts ...float64) []float64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []float64{}
	}
	tmp := make(map[float64]struct{})
	for _, v := range b._set(value[0]) {
		if pattern_float.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be float set", key)
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be float set", key)
		}
		tmp[f] = struct{}{}
	}
	if len(tmp) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []float64{}
	}
	ret := make([]float64, 0, len(tmp))
	for k, _ := range tmp {
		ret = append(ret, k)
	}
	return ret
}

func (b *base) xFloatSet(key string) []float64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	tmp := make(map[float64]struct{})
	for _, v := range b._set(value[0]) {
		if pattern_float.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be float set", key)
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be float set", key)
		}
		tmp[f] = struct{}{}
	}
	if len(tmp) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret := make([]float64, 0, len(tmp))
	for k, _ := range tmp {
		ret = append(ret, k)
	}
	return ret
}

func (b *base) oStrArray(key string, opts ...string) []string {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []string{}
	}
	ret := b._array(value[0])
	if len(ret) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []string{}
	}
	return ret
}

// string array, do not support "" element
func (b *base) xStrArray(key string) []string {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	ret := b._array(value[0])
	if len(ret) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	return ret
}

func (b *base) oIntArray(key string, opts ...int64) []int64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []int64{}
	}
	tmp := b._array(value[0])
	if len(tmp) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []int64{}
	}
	ret := make([]int64, 0, len(tmp))
	for _, v := range tmp {
		if pattern_int.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be integer array", key)
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be integer array", key)
		}
		ret = append(ret, i)
	}
	return ret
}

func (b *base) xIntArray(key string) []int64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	tmp := b._array(value[0])
	if len(tmp) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret := make([]int64, 0, len(tmp))
	for _, v := range tmp {
		if pattern_int.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be integer array", key)
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be integer array", key)
		}
		ret = append(ret, i)
	}
	return ret
}

func (b *base) oUintArray(key string, opts ...uint64) []uint64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []uint64{}
	}
	tmp := b._array(value[0])
	if len(tmp) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []uint64{}
	}
	ret := make([]uint64, 0, len(tmp))
	for _, v := range tmp {
		if pattern_uint.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be unsigned integer array", key)
		}
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be unsigned integer array", key)
		}
		ret = append(ret, i)
	}
	return ret
}

func (b *base) xUintArray(key string) []uint64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	tmp := b._array(value[0])
	if len(tmp) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret := make([]uint64, 0, len(tmp))
	for _, v := range tmp {
		if pattern_uint.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be unsigned integer array", key)
		}
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be unsigned integer array", key)
		}
		ret = append(ret, i)
	}
	return ret
}

func (b *base) oFloatArray(key string, opts ...float64) []float64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []float64{}
	}
	tmp := b._array(value[0])
	if len(tmp) == 0 {
		if len(opts) > 0 {
			return opts
		}
		return []float64{}
	}
	ret := make([]float64, 0, len(tmp))
	for _, v := range tmp {
		if pattern_float.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be float array", key)
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"optional arg[%s], invalid format, should be float array", key)
		}
		ret = append(ret, f)
	}
	return ret
}

func (b *base) xFloatArray(key string) []float64 {
	value, exists := b.req.Form[key]
	if !exists || len(value) == 0 {
		b.exit(ERR_ARG_NOTFOUND, "required arg[%s] not found", key)
	}
	tmp := b._array(value[0])
	if len(tmp) == 0 {
		b.exit(ERR_ARG_EMPTY, "required arg[%s], should not be empty", key)
	}
	ret := make([]float64, 0, len(tmp))
	for _, v := range tmp {
		if pattern_float.MatchString(v) == false {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be float array", key)
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			b.exit(ERR_ARG_INVALID,
				"required arg[%s], invalid format, should be float array", key)
		}
		ret = append(ret, f)
	}
	return ret
}
