//  v0.9.3
// Author: DIEHL E.
// (C) Sony Pictures Entertainment, Jan 2021

package blockchain

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Remote_Client_Invoke(t *testing.T) {

	c, err := NewClient("config", "testdata")
	require.NoError(t, err)
	defer c.Close()

	_, err = c.Invoke("createCar", fmt.Sprintf("CAR%d", rand.Intn(1000)+10), "VW", "Polo", "Grey", "Mary")
	require.NoError(t, err)

}

func Test_Remote_Client_Query(t *testing.T) {
	c, err := NewClient("config", "testdata")
	require.NoError(t, err)
	defer c.Close()

	_, err = c.Query("queryAllCars")
	require.NoError(t, err)
}
