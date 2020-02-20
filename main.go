/*
MIT License

Copyright (c) 2018 iota-tangle.io

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
    "log"
    "math/rand"
    "strings"
    "time"

    "github.com/iotaledger/iota.go/api"
    "github.com/iotaledger/iota.go/address"
    "github.com/iotaledger/iota.go/bundle"
    "github.com/iotaledger/iota.go/trinary"
    "github.com/iotaledger/iota.go/pow"
    flag "github.com/ogier/pflag"
    "github.com/paulbellamy/ratecounter"
)

var (
    randomTag     string
    randomAddress string
    randomSeed    string
    alphabet      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
    rand.Seed(time.Now().UnixNano())

    // Generate random tag suffix to distinguish instances
    for i := 0; i < 3; i++ {
        randomTag += string(alphabet[rand.Intn(len(alphabet))])
    }

    // Send to a random address for easy confirmation checking on thetangle
    alphabet := alphabet + "9"
    for i := 0; i < 81; i++ {
        randomAddress += string(alphabet[rand.Intn(len(alphabet))])
    }

    for i := 0; i < 81; i++ {
        randomSeed += string(alphabet[rand.Intn(len(alphabet))])
    }
}

func main() {
    var mwm *uint64 = flag.Uint64("mwm", 14, "minimum weight magnitude")
    var depth *uint64 = flag.Uint64("depth", 1, "depth for tip finding")
    var destAddress *string = flag.String("address", "<random>", "address to send to")
    var tag *string = flag.String("tag", "999GOPOW9<pow>9<random>", "transaction tag")
    var server *string = flag.String("node", "http://localhost:14265", "remote node to connect to")
    flag.Parse()

    // Set a random address if none was specified
    if *destAddress == "<random>" {
        *destAddress = randomAddress
    }
    recipientT := trinary.Trytes(*destAddress)

    log.Println("Using IRI server:", *server)

    name, powFunc := pow.GetFastestProofOfWorkImpl()

    endpointAPI, _ := api.ComposeAPI(api.HTTPClientSettings{ URI: *server, LocalProofOfWorkFunc: powFunc })

    // Set a random tag if the user didnt specify one
    if *tag == "999GOPOW9<pow>9<random>" {
        *tag = "999GOPOW9" + strings.ToUpper(name) + "9" + randomTag
    }

    ttag := trinary.Trytes(*tag)

    recipientChecksum, _ := address.Checksum(recipientT)

    trs := bundle.Transfers{
    {
            Address: recipientT + recipientChecksum,
            Value:   0,
            Tag:     ttag,
        },
    }

    log.Println("Using tag: http://thetangle.org/tag/" + *tag)
    log.Println("Using address: http://thetangle.org/address/" + *destAddress)
    log.Println("Using PoW:", name)

    // Setup 1/5/15 minue average TPS counters
    r1 := ratecounter.NewRateCounter(1 * time.Minute)
    r5 := ratecounter.NewRateCounter(5 * time.Minute)
    r15 := ratecounter.NewRateCounter(15 * time.Minute)
    prepTransferOpts := api.PrepareTransfersOptions{}
    for {
        trytes, err := endpointAPI.PrepareTransfers(randomSeed, trs, prepTransferOpts)
        if err != nil {
            log.Println("Error preparing transfer:", err)
            continue
        }

        // This is where the PoW is done right before sending the txn
        myBundle, err := endpointAPI.SendTrytes(trytes, *depth, *mwm)
        if err != nil {
            log.Println("Error sending trytes:", err)
            continue
        }

        // Increment counters
        r1.Incr(1)
        r5.Incr(1)
        r15.Incr(1)

        log.Println("SENT: http://thetangle.org/transaction/" + myBundle[0].Hash)
        // 1/5/15 min TPS averages
        log.Printf("TPS: %.3f %.3f %.3f\n", float64(r1.Rate())/float64(60), float64(r5.Rate())/float64(60*5), float64(r15.Rate())/float64(60*15))
    }
}
