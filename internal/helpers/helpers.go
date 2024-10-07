package helpers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

func Deref[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

func MinDistance(word1a, word2a string) int {
	word1 := []rune(word1a)
	word2 := []rune(word2a)

	pre := make([]int, len(word2)+1)
	cur := make([]int, len(word2)+1)
	for i := 0; i < len(pre); i++ {
		pre[i] = i
	}
	for i := 1; i <= len(word1); i++ {
		cur[0] = i
		for j := 1; j < len(pre); j++ {
			if word1[i-1] != word2[j-1] {
				cur[j] = MinOf(cur[j-1], pre[j-1], pre[j]) + 1
			} else {
				cur[j] = pre[j-1]
			}
		}
		tmp := make([]int, len(cur))
		copy(tmp, cur)
		pre = tmp
	}
	return pre[len(word2)]
}

func MinOf(nums ...int) int {
	ans := nums[0]
	for _, v := range nums {
		if v < ans {
			ans = v
		}
	}
	return ans
}

func Map[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item, i)
	}

	return result
}

func Ptr[T any](a T) *T {
	return &a
}

func Empty[T any](a T, found bool) *T {
	if !found {
		return nil
	}

	return &a
}

func UUID() string {
	id := uuid.New()

	return id.String()
}

func UID() uuid.UUID {
	id := uuid.New()

	return id
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func InArray[T comparable](s T, arr []T) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}

	return false
}

func Join(arr []string, del string) string {
	return strings.Join(arr, del)
}

func DateNow() string {
	return time.Now().Format(time.RFC3339)
}

func DateNowMilli() int64 {
	return time.Now().UnixMilli()
}

func DateNowNanosecond() int {
	return time.Now().Nanosecond()
}

func Default[T any](v *T, def T) T {
	if v == nil {
		return def
	}

	return *v
}

func IsValidUUID(u string) bool {
	if len(u) != 36 {
		return false
	}

	_, err := uuid.Parse(u)
	return err == nil
}

//

func RandomPartFromSlice[T any](arr []T) (bool, []T) {
	var result []T

	if len(arr) == 0 {
		return false, result
	}

	for i := 0; i < len(arr); i++ {
		if i%3 == 0 {
			result = append(result, arr[i])
		}
	}

	return true, result
}

func RandomFromSlice[T any](arr []T) (bool, T) {
	var result T

	if len(arr) == 0 {
		return false, result
	}

	if len(arr) == 1 {
		return false, arr[0]
	}

	//nolint // it is ok
	idx := rand.Intn(len(arr))

	return true, arr[idx]
}

func RandomNumber(min, max int) int {
	//nolint // it is ok
	return rand.Intn(max) + min
}

func RandomBigNumber() int {
	return RandomNumber(0, 1000000)
}

func UUIDByHash(s string) string {
	id := uuid.NewSHA1(uuid.UUID{}, []byte(s))

	return id.String()
}

func UUIDByTwoStrings(s1, s2 string) string {
	sorted := ""
	if s1 < s2 {
		sorted = s1 + "." + s2
	} else {
		sorted = s2 + "." + s1
	}

	id := uuid.NewSHA1(uuid.UUID{}, []byte(sorted))

	return id.String()
}

func GetType(s interface{}) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

//

// Return sorted keys.
func SortMapByKeys[T constraints.Ordered](mp map[T]any) []T {
	keys := make([]T, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}

	lenCmp := func(a, b T) int {
		if a < b {
			return -1
		}

		return 1
	}

	slices.SortFunc(keys, lenCmp)

	return keys
}

//

func RemoveTagsFromString(s string) string {
	ind := strings.LastIndex(s, "]") + 1

	return strings.Trim(s[ind:], " ")
}

//

func ConvertPostgresCreds(creds string) (string, error) {
	logrus.Debug(creds)
	// parse string: postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable

	pattern := regexp.MustCompile(`postgres://(?P<user>[^:]+):(?P<password>[^@]+)@(?P<host>[^:]+):(?P<port>[^/]+)/(?P<dbname>[^?]+)(\?sslmode=(?P<sslmode>[^&]+)&sslrootcert=(?P<sslrootcert>[^&]+))*`)

	sub := pattern.FindStringSubmatch(creds)
	connStr := ""
	if len(sub) < 6 {
		logrus.Debug("postgres creds", sub)
		return "", fmt.Errorf("invalid postgres connection string")
	}

	user := sub[1]
	password := sub[2]
	host := sub[3]

	port, err := strconv.Atoi(sub[4])
	if err != nil {
		return "", err
	}

	dbName := sub[5]

	//

	sslMode := sub[7]
	sslCert := sub[8]

	if sslCert != "" {
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s sslrootcert=%s",
			host, port, user, password, dbName, sslMode, sslCert)
	} else {
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName)
	}

	return connStr, nil
}

func Hash(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return string(hashedPassword)
}

func VerifyHash(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func RemoteTagsFromError(msg string) string {
	l := 0

	if msg[0] == '[' {
		for i, c := range msg {
			if i-1 != len(msg) && c == ']' && msg[i+1] == ' ' {
				l = i
				break
			}
		}

		return msg[l+2:]
	}

	return msg
}

func RandomCode(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, length)
	for i := range b {
		//nolint // it is ok
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandomNumCode(length int) string {
	letters := []rune("123456789")

	b := make([]rune, length)
	for i := range b {
		//nolint // it is ok
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateValidationSimpleCode() string {
	return RandomNumCode(6)
}

func GenerateValidationCode() string {
	return RandomNumCode(20)
}

func GenerateResetCode() string {
	return RandomCode(20)
}

func ToSnake(str string) string {
	return strcase.ToSnake(str)
}

func ToLowerSnake(str string) string {
	return strings.ToLower(strcase.ToSnake(str))
}

func ParsePathFileName(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

func ParsePathExt(path string) string {
	return filepath.Ext(path)
}

func ParsePathBase(path string) string {
	return filepath.Base(path)
}

func PathInsertSize(path string, size int) string {
	if size == 0 {
		return path
	}

	fileName := ParsePathFileName(path)
	ext := filepath.Ext(path)

	return fmt.Sprintf("%s.w%d%s", fileName, size, ext)
}

func IntToLetters(number int) (letters string) {
	number--

	if firstLetter := number / 26; firstLetter > 0 {
		letters += IntToLetters(firstLetter)
		letters += string(rune('a' + number%26))
	} else {
		letters += string(rune('a' + number))
	}

	return letters
}

func GetMapKeys[T comparable](mymap map[T]interface{}) []T {
	keys := make([]T, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}

	return keys
}

func ArrayIntersection(a1, a2 []string) []string {
	res := []string{}

	longestArr := a1
	shortestArr := a2
	if len(a1) < len(a2) {
		longestArr = a2
		shortestArr = a1
	}

	keys := make(map[string]bool)

	for _, k := range shortestArr {
		keys[k] = true
	}

	for _, k := range longestArr {
		if _, ok := keys[k]; ok {
			res = append(res, k)
		}
	}

	return res
}

func ArrayNonIntersection(a1, a2 []string) []string {
	res := []string{}

	longestArr := a1
	shortestArr := a2
	if len(a1) < len(a2) {
		longestArr = a2
		shortestArr = a1
	}

	keys := make(map[string]bool)

	for _, k := range shortestArr {
		keys[k] = true
	}

	for _, k := range longestArr {
		if _, ok := keys[k]; !ok {
			res = append(res, k)
		}
	}

	return res
}

func FileMimetype(path string) (string, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return "", err
	}

	return mtype.String(), nil
}

func FileIsImage(path string) (bool, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return false, err
	}

	return FileMimeIsImage(mtype.String()), nil
}

func FileMimeIsImage(mime string) bool {
	return mimetype.EqualsAny(mime, "image/jpeg", "image/png", "image/tiff")
}

func FileMimeToPreview(mime string) bool {
	return mimetype.EqualsAny(mime, "image/jpeg", "image/png", "image/tiff", "image/webp", "image/gif")
}

func FileSize(path string) (int64, error) {
	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return f.Size(), nil
}

func FileExt(path string) string {
	return filepath.Ext(path)
}

func MustInt(i string) int {
	v, err := strconv.Atoi(i)
	if err != nil {
		return -1
	}

	return v
}

func PatchPath(parent, current string, path []string) []string {
	if len(path) == 0 {
		return lo.Uniq([]string{parent, current})
	}

	res := []string{}
	for _, a := range lo.Uniq(path) {
		if a == current {
			res = append(res, parent)
		}

		if a != parent {
			res = append(res, a)
		}
	}

	return res
}

func Unique[T comparable](arr []T) []T {
	return lo.Uniq(arr)
}

// Alloc uint64
// Alloc is bytes of allocated heap objects.
// "Allocated" heap objects include all reachable objects, as well as unreachable objects that the garbage collector has not yet freed.
// Specifically, Alloc increases as heap objects are allocated and decreases as the heap is swept and unreachable objects are freed.
// Sweeping occurs incrementally between GC cycles, so these two processes occur simultaneously, and as a result Alloc tends to change smoothly (in contrast with the sawtooth that is typical of stop-the-world garbage collectors).
//
// TotalAlloc uint64
// TotalAlloc is cumulative bytes allocated for heap objects.
// TotalAlloc increases as heap objects are allocated, but unlike Alloc and HeapAlloc, it does not decrease when objects are freed.
//
// Sys uint64
// Sys is the total bytes of memory obtained from the OS.
// Sys is the sum of the XSys fields below. Sys measures the virtual address space reserved by the Go runtime for the heap, stacks, and other internal data structures. It's likely that not all of the virtual address space is backed by physical memory at any given moment, though in general it all was at some point.
//
// NumGC uint32
// NumGC is the number of completed GC cycles.

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func ToInterfaceMap[T any](mp map[string]T) map[string]interface{} {
	res := make(map[string]interface{})

	for k, v := range mp {
		res[k] = v
	}

	return res
}

func EquelSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func StructToMap(obj interface{}) (newMap map[string]interface{}, err error) {
	data, err := json.Marshal(obj) // Convert to a json string
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &newMap) // Convert to a map
	return newMap, err
}

func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func ToInterface[T any](in []T) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func FindNewElements[T comparable](oldArray, newArray []T) []T {
	res := []T{}
	for _, n := range newArray {
		if lo.IndexOf(oldArray, n) == -1 {
			res = append(res, n)
		}
	}

	return res
}

func FindRemovedElements[T comparable](oldArray, newArray []T) []T {
	res := []T{}
	for _, n := range oldArray {
		if lo.IndexOf(newArray, n) == -1 {
			res = append(res, n)
		}
	}

	return res
}
