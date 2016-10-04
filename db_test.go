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
)

var testDB *DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = NewDB("user=tigerblood dbname=tigerblood sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer testDB.Close()
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
	err := testDB.emptyReputationTable()
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
	err := testDB.emptyReputationTable()
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
	assert.Nil(t, testDB.emptyReputationTable())
	assert.Nil(t, testDB.InsertReputationEntry(nil, ReputationEntry{IP: "192.168.0.1", Reputation: 0}))
	assert.Nil(t, testDB.UpdateReputationEntry(nil, ReputationEntry{IP: "192.168.0.1", Reputation: 1}))
	entry, err := testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(1), entry.Reputation)
}

func TestDelete(t *testing.T) {
	assert.Nil(t, testDB.emptyReputationTable())
	assert.Nil(t, testDB.InsertReputationEntry(nil, ReputationEntry{IP: "192.168.0.1", Reputation: 0}))
	assert.Nil(t, testDB.DeleteReputationEntry(nil, ReputationEntry{IP: "192.168.0.1"}))
	_, err := testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.NotNil(t, err)
}
