package main
 
import (
    "fmt"
    "testing"
    "strings"
    "strconv"
    "math/big"
)
 
/*
// test generator function
func TestCheckDHPair(t *testing.T) {
    //add a few to use non-2 generators
    gentest := []string{"541:2",
			"440527:2",
			"506119:2",
                       }
    testint := big.NewInt(1)
    for _, pair := range(gentest) {
	fmt.Println("Working on pair ", pair)
	testint.SetString(strings.Split(pair, ":")[0], 10) 
	output := makeGenerator(testint)
	expectedOutput, err := strconv.Atoi(strings.Split(pair, ":")[1])
	if err != nil {
	    panic(fmt.Sprintf("Error in string conversion: %q", err))
        }
	if output != expectedOutput {
	    t.Errorf("Expected Int(%d) is not same as actual Int(%d)", expectedOutput, output)
	}
    }
}

func TestCheckPrimeNumber(t *testing) {
}

func TestAppendIfMissing(t *testing) {
}

func TestFindPrimeFactors(t *testing) {
}

func TestPrimRootCheck(t *testing) {
}

func TestCheckGenerator(t *testing) {
}

func TestCheckPrivkey(t *testing) {
}

*/
