package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
    // prepare
    db, err := sql.Open("sqlite", "tracker.db")
    if err != nil {
        require.NoError(t, err)
    }
    defer db.Close()

    store := NewParcelStore(db)
    parcel := getTestParcel()

    // add
    // добавляем новую посылку в БД, проверяем отсутствие ошибки и наличии идентификатора
    id, err := store.Add(parcel)
    require.NoError(t, err)
    require.Positive(t, id)
    parcel.Number = id

    // get
    // получаем только что добавленную посылку, проверяем отсутствие ошибки
    // проверяем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
    stored, err := store.Get(id)
    require.NoError(t, err)
    assert.Equal(t, parcel, stored)

    // delete
    err = store.Delete(id) 
    require.NoError(t, err)

    _, err = store.Get(id)
    require.Error(t, err)
    assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
    // prepare

    db, err := sql.Open("sqlite", "tracker.db")
    if err != nil {
        require.NoError(t, err)
    }
    defer db.Close()

    store := NewParcelStore(db)
    parcel := getTestParcel()

    // add
    // добавляем новую посылку в БД, проверяем отсутствие ошибки и наличии идентификатора

    id, err := store.Add(parcel)
    require.NoError(t, err)
    require.NotEmpty(t, id)

    // set address
    // обновляем адрес, проверяем в отсутствии ошибки
    newAddress := "new test address"
    err = store.SetAddress(parcel.Number, newAddress)
    require.NoError(t, err)

    // check
    // добавляем посылку и проверяем, что адрес обновился
    stored, err := store.Get(id)
    require.NoError(t, err)
    assert.Equal(t, parcel.Address, stored.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
    // prepare

    db, err := sql.Open("sqlite", "tracker.db")
    if err != nil {
        require.NoError(t, err)
    }
    defer db.Close()

    store := NewParcelStore(db)
    parcel := getTestParcel()

    // add
    // добавляем новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
    id, err := store.Add(parcel)
    require.NoError(t, err)
    require.NotEmpty(t, id)

    // set status

    var nextStatus string
    switch parcel.Status {
    case ParcelStatusRegistered:
        nextStatus = ParcelStatusSent
    case ParcelStatusSent:
        nextStatus = ParcelStatusDelivered
    }

    err = store.SetStatus(id, nextStatus)
    require.NoError(t, err)

    // check
    // получаем добавленную посылку и убедитесь, что статус обновился
    stored, err := store.Get(id)
    require.NoError(t, err)
    assert.Equal(t, nextStatus, stored.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
    // prepare

    db, err := sql.Open("sqlite", "tracker.db")
    if err != nil {
        require.NoError(t, err)
    }
    defer db.Close()

    store := NewParcelStore(db)

    parcels := []Parcel{
        getTestParcel(),
        getTestParcel(),
        getTestParcel(),
    }


    // задаём всем посылкам один и тот же идентификатор клиента
    client := randRange.Intn(10_000_000)
    parcels[0].Client = client
    parcels[1].Client = client
    parcels[2].Client = client

    // add
    for i := 0; i < len(parcels); i++ {
        id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
        require.NoError(t, err)
        require.NotEmpty(t, id)

        // обновляем идентификатор добавленной у посылки
        parcels[i].Number = id
    }

    // get by client
    storedParcels, err := store.GetByClient(client)
    require.NoError(t, err)
    assert.ElementsMatch(t, parcels, storedParcels)
}

