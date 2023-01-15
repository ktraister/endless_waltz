package main

import (
    "fmt"
    "math/big"
)

func primesLessThan(n *big.Int) (primes []big.Int) {
    var one big.Int
    one.SetInt64(1)
    var i big.Int
    i.SetInt64(1)
    for i.Cmp(n) < 0 {
        var result big.Int
        result.Set(&i)
        fmt.Println(result.String())
        primes = append(primes, result)
        i.Add(&i, &one)
    }
    return
}

func main() {
    primes := primesLessThan(big.NewInt(10))
    for _, p := range primes {
        fmt.Println(p.String())
    }
}
