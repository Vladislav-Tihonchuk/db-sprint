package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel.Client, storedParcel.Client)
	assert.Equal(t, parcel.Status, storedParcel.Status)
	assert.Equal(t, parcel.Address, storedParcel.Address)
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, storedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, storedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	for _, stored := range storedParcels {
		original, found := parcelMap[stored.Number]
		assert.True(t, found)
		assert.Equal(t, original, stored)
		assert.Equal(t, original.Status, stored.Status)
		assert.Equal(t, original.Address, stored.Address)
	}
}

func TestDeleteWithIncorrectStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)
	err = store.Delete(id)
	require.Error(t, err)
	assert.Equal(t, "нельзя удалить посылку со статусом sent", err.Error())
}
