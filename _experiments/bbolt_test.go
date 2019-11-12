package bench_bolt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"go.etcd.io/bbolt"
)

var working_Directory string
var bolt_db *bbolt.DB
var bolt_bucketName = []byte("scs:session")

func init() {
	var err error

	working_Directory = os.TempDir() + "\\experiments"
	err = os.MkdirAll(working_Directory, 0664)
	if err != nil {
		fmt.Println("Unable to create session folder!")
		panic(err)
	}
	fmt.Printf("%-32s %s\n", "| Working Directory:", working_Directory)

	createBoltDB()
}

func generateTestValue(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func createBoltDB() {
	var err error

	if bolt_db != nil {
		_ = bolt_db.Close()
	}

	os.Remove(working_Directory + "\\bolt.db")

	// Establish a connection to Bolt database
	bolt_db, err = bbolt.Open(working_Directory+"\\bolt.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	err = bolt_db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bolt_bucketName)
		return err
	})
}

func insert(n int, size int) {
	err := bolt_db.Update(func(tx *bbolt.Tx) error {
		var err error

		var token string
		token = fmt.Sprintf("%032d", n)

		var buf string
		buf, err = generateTestValue(size)

		bucket := tx.Bucket(bolt_bucketName)
		err = bucket.Put([]byte(token), []byte(buf))
		return err
	})

	if err != nil {
		log.Fatalln(err)
	}
}
func insert_loop(b *testing.B, size int) {
	for n := 1; n <= b.N; n++ {
		insert(n, size)
	}
}

var sm = &sync.Mutex{}
var sn int

func insert_paralel(b *testing.B, size int) {
	sn = 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sm.Lock()
			n := sn
			sn++
			sm.Unlock()

			insert(n, size)
		}
	})
}

func BenchmarkBoltInsert_128(b *testing.B) {
	createBoltDB()
	b.ResetTimer()

	insert_loop(b, 128)
}
func BenchmarkBoltInsert_1024(b *testing.B) {
	createBoltDB()
	b.ResetTimer()

	insert_loop(b, 1024)
}
func BenchmarkBoltInsertParalel_128(b *testing.B) {
	createBoltDB()
	b.ResetTimer()

	insert_paralel(b, 128)
}
func BenchmarkBoltInsertParalel_1024(b *testing.B) {
	createBoltDB()
	b.ResetTimer()

	insert_paralel(b, 1024)
}

func find(n int) {
	var val []byte
	bolt_db.View(func(tx *bbolt.Tx) error {

		var token string
		token = fmt.Sprintf("%032d", n)

		bucket := tx.Bucket(bolt_bucketName)
		val = bucket.Get([]byte(token))

		return nil
	})

	if val == nil {
		log.Fatalln("not found!", n)
	}
}
func find_loop(b *testing.B) {
	for n := 1; n <= b.N; n++ {
		find(n)
	}
}
func find_paralel(b *testing.B) {
	sn = 1

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sm.Lock()
			n := sn
			sn++
			sm.Unlock()

			find(n)
		}
	})
}

func BenchmarkBoltFind_128(b *testing.B) {
	createBoltDB()
	insert_loop(b, 128)
	b.ResetTimer()

	find_loop(b)
}
func BenchmarkBoltFind_1024(b *testing.B) {
	createBoltDB()
	insert_loop(b, 1024)
	b.ResetTimer()

	find_loop(b)
}

func BenchmarkBoltFindParalel_128(b *testing.B) {
	createBoltDB()
	insert_loop(b, 128)
	b.ResetTimer()

	find_paralel(b)
}
func BenchmarkBoltFindParalel_1024(b *testing.B) {
	createBoltDB()
	insert_loop(b, 1024)
	b.ResetTimer()

	find_paralel(b)
}
