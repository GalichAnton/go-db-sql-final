package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

func TestParcelStore(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	tests := []struct {
		name     string
		testFunc func(*testing.T, *ParcelStore)
	}{
		{
			name: "AddGetDelete",
			testFunc: func(t *testing.T, store *ParcelStore) {
				parcel := getTestParcel()

				id, err := store.Add(parcel)
				require.NoError(t, err)
				require.NotEmpty(t, id)

				p, err := store.Get(id)
				require.NoError(t, err)
				require.Equal(t, id, p.Number)
				parcel.Number = id
				require.Equal(t, parcel, p)

				err = store.Delete(id)
				require.NoError(t, err)
				_, err = store.Get(id)
				require.Error(t, err)
			},
		},
		{
			name: "SetAddress",
			testFunc: func(t *testing.T, store *ParcelStore) {
				parcel := getTestParcel()

				id, err := store.Add(parcel)
				require.NoError(t, err)
				require.NotEmpty(t, id)

				newAddress := "new test address"
				err = store.SetAddress(id, newAddress)
				require.NoError(t, err)

				p, err := store.Get(id)
				require.NoError(t, err)
				require.Equal(t, newAddress, p.Address)
			},
		},
		{
			name: "TestSetStatus",
			testFunc: func(t *testing.T, store *ParcelStore) {
				parcel := getTestParcel()

				id, err := store.Add(parcel)
				require.NoError(t, err)
				require.NotEmpty(t, id)

				newStatus := ParcelStatusSent
				err = store.SetStatus(id, newStatus)
				require.NoError(t, err)

				p, err := store.Get(id)
				require.NoError(t, err)
				require.Equal(t, newStatus, p.Status)
			},
		},
		{
			name: "TestGetByClient",
			testFunc: func(t *testing.T, store *ParcelStore) {
				parcels := []Parcel{
					getTestParcel(),
					getTestParcel(),
					getTestParcel(),
				}
				parcelMap := map[int]Parcel{}

				// задаём всем посылкам один и тот же идентификатор клиента
				client := randRange.Intn(10_000_000)
				parcels[0].Client = client
				parcels[1].Client = client
				parcels[2].Client = client

				for i, parcel := range parcels {
					id, err := store.Add(parcel)
					require.NoError(t, err)
					require.NotEmpty(t, id)

					parcels[i].Number = id

					parcelMap[id] = parcels[i]
				}

				storedParcels, err := store.GetByClient(client)
				require.NoError(t, err)
				require.Equal(t, len(parcels), len(storedParcels))

				for _, parcel := range storedParcels {
					expectedParcel, exists := parcelMap[parcel.Number]
					require.True(t, exists)
					require.Equal(t, expectedParcel, parcel)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t, &store)
		})
	}
}
