package tigerblood

import (
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"
)

var testDB *DB

func TestMain(m *testing.M) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	if found != true {
		log.Fatal("TIGERBLOOD_DSN not found in test env.")
	}

	var err error
	testDB, err = NewDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer testDB.Close()
	exceptionTestHook = true
	os.Exit(m.Run())
}

func TestCreateSchema(t *testing.T) {
	err := testDB.CreateTables()
	assert.Nil(t, err)
	err = testDB.CreateTables()
	assert.Nil(t, err, "Running CreateTables when the tables already exist shouldn't error")
}

func TestReputationInsertConstraint(t *testing.T) {
	err := testDB.InsertReputationEntry(nil, ReputationEntry{IP: "240.0.0.1", Reputation: 500})
	assert.IsType(t, CheckViolationError{}, err)
	err = testDB.InsertReputationEntry(nil, ReputationEntry{IP: "240.0.0.1", Reputation: 50})
	assert.Nil(t, err)
}

func TestReputationUpdateConstraint(t *testing.T) {
	err := testDB.UpdateReputationEntry(nil, ReputationEntry{IP: "240.0.0.1", Reputation: 500})
	assert.IsType(t, CheckViolationError{}, err)
	err = testDB.UpdateReputationEntry(nil, ReputationEntry{IP: "240.0.0.1", Reputation: 50})
	assert.Nil(t, err)
}

func randomCidr(minSubnet, maxSubnet uint) string {
	// Get a subnet with mean on 24 and a stdev of 5
	subnet := math.Abs(rand.NormFloat64())*24 + 5
	subnet = math.Min(subnet, float64(maxSubnet))
	// The biggest subnets will be /8s.
	subnet = math.Max(subnet, float64(minSubnet))
	ip := rand.Uint32()
	netmask := make([]byte, 4)
	binary.BigEndian.PutUint32(netmask, ^((1 << (32 - uint(subnet))) - 1))
	ipBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(ipBytes, ip)
	for i := range ipBytes {
		ipBytes[i] &= netmask[i]
	}
	return fmt.Sprintf("%d.%d.%d.%d/%d", uint(ipBytes[0]), uint(ipBytes[1]), uint(ipBytes[2]), uint(ipBytes[3]), uint(subnet))
}

func BenchmarkInsertion(b *testing.B) {
	err := testDB.EmptyTables()
	assert.Nil(b, err)
	b.RunParallel(func(pb *testing.PB) {
		var ip [1000]string
		generateRandomIps := func() {
			b.StopTimer()
			for i := range ip {
				ip[i] = randomCidr(8, 32)
			}
			b.StartTimer()
		}

		generateRandomIps()
		for i := 0; pb.Next(); i++ {
			if i%1000 == 0 {
				generateRandomIps()
			}
			currIP := ip[i%1000]
			err := testDB.InsertOrUpdateReputationEntry(nil, ReputationEntry{
				IP:         currIP,
				Reputation: 50,
			})
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkSelection(b *testing.B) {
	err := testDB.EmptyTables()
	assert.Nil(b, err)
	tx, err := testDB.Begin()
	assert.Nil(b, err)
	var ip [1000]string
	for i := 0; i < 10000; i++ {
		if i%1000 == 0 {
			for j := range ip {
				ip[j] = randomCidr(8, 32)
			}
		}
		currIP := ip[i%1000]
		err := testDB.InsertOrUpdateReputationEntry(tx, ReputationEntry{
			IP:         currIP,
			Reputation: 50,
		})
		if err != nil {
			b.Error(err)
		}
	}
	tx.Commit()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testDB.SelectSmallestMatchingSubnet(randomCidr(32, 32))
		}
	})
}

func TestUpdate(t *testing.T) {
	assert.Nil(t, testDB.EmptyTables())
	assert.NotNil(t, testDB.UpdateReputationEntry(nil, ReputationEntry{IP: "192.168.0.1", Reputation: 1}))
	assert.Nil(t, testDB.InsertReputationEntry(nil, ReputationEntry{IP: "192.168.0.1", Reputation: 0}))
	assert.Nil(t, testDB.UpdateReputationEntry(nil, ReputationEntry{IP: "192.168.0.1", Reputation: 1}))
	entry, err := testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(1), entry.Reputation)
}

func TestDelete(t *testing.T) {
	assert.Nil(t, testDB.EmptyTables())
	assert.Nil(t, testDB.InsertReputationEntry(nil, ReputationEntry{IP: "192.168.0.1", Reputation: 0}))
	assert.Nil(t, testDB.DeleteReputationEntry(nil, ReputationEntry{IP: "192.168.0.1"}))
	_, err := testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.NotNil(t, err)
}

func TestInsertOrUpdateReputationPenalties(t *testing.T) {
	assert.Nil(t, testDB.CreateTables())
	assert.Nil(t, testDB.EmptyTables())

	// test insert
	err := testDB.InsertOrUpdateReputationPenalties(nil, []string{"192.168.0.1"}, []uint{90})
	assert.Nil(t, err)

	entry, err := testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(10), entry.Reputation)

	// test update
	err = testDB.InsertOrUpdateReputationPenalties(nil, []string{"192.168.0.1"}, []uint{9})
	assert.Nil(t, err)

	entry, err = testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(1), entry.Reputation)

	// test reputation doesn't go negative
	err = testDB.InsertOrUpdateReputationPenalties(nil, []string{"192.168.0.1"}, []uint{90})
	assert.Nil(t, err)

	entry, err = testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(0), entry.Reputation)

	assert.Nil(t, testDB.EmptyTables())
}

func TestExceptionUpdate(t *testing.T) {
	assert.Nil(t, testDB.EmptyTables())
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.5.0/24",
		Creator: "file:/test",
	}))
	ret, err := testDB.SelectExceptionsContaining("10.0.0.5")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ret))
	ret, err = testDB.SelectExceptionsContaining("10.0.5.5")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.0.0/8",
		Creator: "file:/test2",
	}))
	ret, err = testDB.SelectExceptionsContaining("10.0.5.10")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ret))
	oldts := ret[1].Modified
	// Add same exception again for update
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.0.0/8",
		Creator: "file:/test2",
	}))
	ret, err = testDB.SelectExceptionsContaining("10.0.5.10")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ret))
	assert.NotEqual(t, oldts, ret[1].Modified)
}

func TestExceptionUpdateBad(t *testing.T) {
	assert.Nil(t, testDB.EmptyTables())
	assert.NotNil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "1.2.3.4/40",
		Creator: "file:/test",
	}))
}

func TestExceptionContainedBy(t *testing.T) {
	assert.Nil(t, testDB.EmptyTables())
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.5.0/24",
		Creator: "file:/test",
	}))
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.6.0/24",
		Creator: "file:/test",
	}))
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "192.168.0.0/16",
		Creator: "file:/test",
	}))
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.7.0/24",
		Creator: "file:/test",
	}))
	ret, err := testDB.SelectExceptionsContainedBy("10.0.5.0/24")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	ret, err = testDB.SelectExceptionsContainedBy("192.0.0.0/8")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	ret, err = testDB.SelectExceptionsContainedBy("10.0.0.0/8")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(ret))
	ret, err = testDB.SelectAllExceptions()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(ret))
}

func TestDeleteExpiredExceptions(t *testing.T) {
	assert.Nil(t, testDB.EmptyTables())
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.0.0/8",
		Creator: "file:/test2",
		Expires: time.Now().Add(-1 * (time.Minute * 60)),
	}))
	ret, err := testDB.SelectExceptionsContaining("10.20.0.50")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	assert.Nil(t, testDB.DeleteExpiredExceptions(nil))
	ret, err = testDB.SelectExceptionsContaining("10.20.0.50")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ret))
}

func TestDeleteExceptionCreatorType(t *testing.T) {
	assert.Nil(t, testDB.EmptyTables())
	assert.Nil(t, testDB.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.0.0/8",
		Creator: "file:/test2",
		Expires: time.Now().Add(-1 * (time.Minute * 60)),
	}))
	ret, err := testDB.SelectExceptionsContaining("10.20.0.50")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	assert.Nil(t, testDB.DeleteExceptionCreatorType(nil, "invalid"))
	ret, err = testDB.SelectExceptionsContaining("10.20.0.50")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	assert.Nil(t, testDB.DeleteExceptionCreatorType(nil, "file"))
	ret, err = testDB.SelectExceptionsContaining("10.20.0.50")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ret))
}
