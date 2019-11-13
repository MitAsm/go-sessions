/*
	go test -bench=Insert -benchmem -benchtime 1000x -count 3 -cpu 1,2,4,8
	go test -bench=Find -benchmem -benchtime 100000x -count 3 -cpu 1,2,4,8
*/

package bench_bolt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	bolt "go.etcd.io/bbolt"
)

var workingDirectory string
var boltDB *bolt.DB
var boltBucket = []byte("sessions")

const storedValueSize int = 256

func init() {
	var err error

	workingDirectory = os.TempDir() + "\\experiments"
	err = os.MkdirAll(workingDirectory, 0664)
	if err != nil {
		fmt.Println("Unable to create session folder!")
		panic(err)
	}
	fmt.Printf("%-32s %s\n", "| Working Directory:", workingDirectory)

}

func createNewDB() {
	var err error

	dropDB()

	// Establish a connection to Bolt database
	boltDB, err = bolt.Open(workingDirectory+"\\bolt.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	// Create bucket
	err = boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(boltBucket)
		return err
	})
}

func dropDB() {
	// Close current connection handle
	if boltDB != nil {
		_ = boltDB.Close()
	}

	// Drop database file
	os.Remove(workingDirectory + "\\bolt.db")
}

// Generate test base64 string value
func generateTestValue(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func insert(n int) {
	err := boltDB.Update(func(tx *bolt.Tx) error {
		var err error

		var token string
		token = fmt.Sprintf("%032d", n)

		var buf string
		buf, err = generateTestValue(storedValueSize)

		bucket := tx.Bucket(boltBucket)
		err = bucket.Put([]byte(token), []byte(buf))
		return err
	})

	if err != nil {
		log.Fatalln(err)
	}
}
func insert_loop(b *testing.B) {
	for n := 1; n <= b.N; n++ {
		insert(n)
	}
}

// parallel variables
var parallelMutex = &sync.Mutex{}
var parallelID int

func insert_parallel(b *testing.B) {
	parallelID = 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			parallelMutex.Lock()
			n := parallelID
			parallelID++
			parallelMutex.Unlock()

			insert(n)
		}
	})
}

func Benchmark_Insert(b *testing.B) {
	createNewDB()
	b.ResetTimer()
	insert_loop(b)
	b.StopTimer()
	dropDB()
}

func Benchmark_InsertParallel(b *testing.B) {
	createNewDB()
	b.ResetTimer()
	insert_parallel(b)
	b.StopTimer()
	dropDB()
}

func find(n int) {
	var token string
	var val []byte
	boltDB.View(func(tx *bolt.Tx) error {

		token = fmt.Sprintf("%032d", n)

		bucket := tx.Bucket(boltBucket)
		val = bucket.Get([]byte(token))

		return nil
	})
	//fmt.Printf("%-32s %s\n", token, val)

	if val == nil {
		log.Fatalln("not found!", n)
	}
}
func find_loop(b *testing.B) {
	for n := 1; n <= b.N; n++ {
		find(n)
	}
}
func find_parallel(b *testing.B) {
	parallelID = 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			parallelMutex.Lock()
			n := parallelID
			parallelID++
			parallelMutex.Unlock()

			find(n)
		}
	})
}

var findInitialized bool = false

func BenchmarkFind(b *testing.B) {
	if findInitialized == false {
		createNewDB()
		insert_loop(b)
	}
	b.ResetTimer()
	find_loop(b)
	//b.StopTimer()
	//dropDB()
}

func BenchmarkFindParallel(b *testing.B) {
	if findInitialized == false {
		createNewDB()
		insert_loop(b)
	}
	b.ResetTimer()
	find_parallel(b)
	//b.StopTimer()
	//dropDB()
}
